package services

import (
	"errors"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mobilemindtech/go-io/option"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-io/rio"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/fault"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"strings"
	"time"
)

type AuthData struct {
	UserName   string `json:"username" valid:"Required"`
	Password   string `json:"password" valid:"Required"`
	TenantUUID string `json:"tenant_uuid"`
}

type AuthUser struct {
	Data *AuthData
	User *models.User
}

type AuthResult struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Type string `json:"type"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RefreshExpiresAt int64  `json:"refresh_expires_at,omitempty"`
}

type AuthService struct {
	UserInfo *models.User
	Session  *db.Session
}

func NewAuthService(session *db.Session) *AuthService {
	return &AuthService{Session: session}
}

func (this *AuthService) IsAuthenticated() bool {
	return db.IsPersisted(this.UserInfo)
}

func (this *AuthService) SetUserInfo(user *models.User) {
	this.UserInfo = user
}

func (this *AuthService) IsAdmin() bool {
	if this.UserInfo != nil && this.UserInfo.Roles != nil {
		for _, it := range *this.UserInfo.Roles {
			if it.Authority == "ROLE_ADMIN" {
				return true
			}
		}
		return this.IsRoot()
	}
	return false
}

func (this *AuthService) IsRoot() bool {
	if this.UserInfo != nil && this.UserInfo.Roles != nil {
		for _, it := range *this.UserInfo.Roles {
			if it.Authority == "ROLE_ROOT" {
				return true
			}
		}
	}

	return false
}

func IsRootUser(user *models.User) bool {
	if user != nil && user.Roles != nil {
		for _, it := range *user.Roles {
			if it.Authority == "ROLE_ROOT" {
				return true
			}
		}
	}

	return false
}

func (this *AuthService) ValidPassword(password1 string, password2 string) error {

	if len(strings.TrimSpace(password1)) == 0 {
		return errors.New("A senha não pode ser vazia")
	}

	if password1 != password2 {
		return errors.New("As senhas devem ser iguais")
	}

	if len(password1) < 4 || len(password1) > 10 {
		return errors.New("O tamanho da senha deve ser de 4 a 10 caracteres")
	}

	return nil
}

func (this *AuthService) GetUserRoles() []string {

	roles := []string{}
	if this.UserInfo != nil && this.UserInfo.Roles != nil {
		for _, it := range *this.UserInfo.Roles {
			roles = append(roles, it.Authority)
		}
	}
	return roles
}

func (this *AuthService) IsBearerToken(token string) bool {
	return strings.HasPrefix(token, "Bearer ")
}

func (this *AuthService) CheckBearerToken(bearerToken string) (*models.User, error) {

	if !this.IsBearerToken(bearerToken) {
		return nil, nil
	}

	token := bearerToken[len("Bearer "):]

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		secret, err := beego.AppConfig.String("jwt_token_secret")
		return []byte(secret), err
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		logs.Error("jwt error:", err)
		return nil, err
	}

	mapClaims := jwtToken.Claims.(jwt.MapClaims)
	uuid, _ := mapClaims["user"].(string)
	expiresAt := support.AnyToInt64(mapClaims["expires_at"])

	// sem uuid a busca abaixo casaria com qualquer usuário de uuid vazio
	if len(strings.TrimSpace(uuid)) == 0 {
		return nil, errors.New("invalid token: empty user")
	}

	logs.Debug("expires_at = %v, now %v", expiresAt, util.DateNow().Unix())

	if expiresAt < util.DateNow().Unix() {
		return nil, errors.New("token expired")
	}

	return criteria.New[*models.User](this.Session).
		Eq("Uuid", uuid).
		First()
}

func (this *AuthService) AuthAdmin(body []byte) *rio.IO[*AuthResult] {
	return this.Auth(body, []string{"ROLE_ADMIN"})
}

func (this *AuthService) AuthRoot(body []byte) *rio.IO[*AuthResult] {
	return this.Auth(body, []string{"ROLE_ROOT"})
}

func (this *AuthService) AuthRootOrAdmin(body []byte) *rio.IO[*AuthResult] {
	return this.Auth(body, []string{"ROLE_ROOT", "ROLE_ADMIN"})
}

// Auth by bearer token
func (this *AuthService) Auth(body []byte, allowedRoles []string) *rio.IO[*AuthResult] {
	return rio.AttemptThen(
		rio.AttemptThenOfIO(this.authenticate(body), this.checkPermissions(allowedRoles, true)),
		func(auth *AuthUser) *result.Result[*AuthResult] {
			return this.IssueTokens(auth.User, "", "")
		})
}

// newAccessToken gera o bearer token (JWT) do usuário com validade ttl
func (this *AuthService) newAccessToken(user *models.User, ttl time.Duration) *result.Result[*AuthResult] {

	if len(strings.TrimSpace(user.Uuid)) == 0 {
		return result.OfError[*AuthResult](fault.NewUserUnauthorizedError("user without uuid"))
	}

	secret, _ := beego.AppConfig.String("jwt_token_secret")
	expiresAt := util.DateNow().Add(ttl).Unix()
	token := support.NewBearerToken(jwt.MapClaims{
		"user":       user.Uuid,
		"expires_at": expiresAt,
	}, secret)

	return result.Map(token, func(token string) *AuthResult {
		return &AuthResult{Token: token, ExpiresAt: expiresAt, Type: "bearer"}
	})
}

// authenticate valida usuário/senha do body (AuthData)
func (this *AuthService) authenticate(body []byte) *rio.IO[*AuthUser] {

	jsonParser := rio.Attempt(func() *result.Result[*AuthData] {
		return json.UnmarshalResult[*AuthData](body)
	})

	entityValidator := rio.AttemptThen(jsonParser, func(auth *AuthData) *result.Result[*AuthData] {
		res := validator.New().
			WithPath("AuthData").
			ValidateOrError(auth)
		return result.MapToValue(res, auth)
	})

	return rio.AttemptThen(entityValidator, func(auth *AuthData) *result.Result[*AuthUser] {

		user := criteria.New[*models.User](this.Session).
			Eq("UserName", auth.UserName).
			Eq("Enabled", true).
			GetFirst()

		return result.FlatMap(
			user,
			func(opt *option.Option[*models.User]) *result.Result[*AuthUser] {

				if opt.IsNone() {
					return result.OfError[*AuthUser](fault.NewUserNotFoundError())
				}

				user := opt.Get()
				
				if !user.IsSamePassword(auth.Password) {
					return result.OfError[*AuthUser](fault.NewUserUnauthorizedError("wrong password"))
				}

				return result.OfValue(&AuthUser{
					auth, user,
				})
			})
	})
}

// checkPermissions valida as roles do usuário e, quando TenantUUID é informado,
// o vínculo dele com a tenant. requireAdmin exige que o vínculo seja de admin
// (clientes de loja são vinculados à tenant sem admin).
func (this *AuthService) checkPermissions(allowedRoles []string, requireAdmin bool) func(*AuthUser) *rio.IO[*AuthUser] {
	return func(auth *AuthUser) *rio.IO[*AuthUser] {
		tenantChecker := rio.Attempt(func() *result.Result[bool] {
			c := criteria.New[*models.TenantUser](this.Session).
				Eq("User", auth.User).
				Eq("Tenant__Uuid", auth.Data.TenantUUID).
				Eq("Enabled", true)
			if requireAdmin {
				c.Eq("Admin", true)
			}
			return c.GetAny()
		}).AttemptThen(func(b bool) *result.Result[bool] {
			if !b {
				return result.OfError[bool](fault.NewUserUnauthorizedError("wrong tenant"))
			}
			return result.OfValue(true)
		})

		roleChecker := rio.Attempt(func() *result.Result[bool] {
			c := criteria.New[*models.UserRole](this.Session).
				Eq("User", auth.User)

			cond := db.NewCondition()
			for _, role := range allowedRoles {
				cond.Eq("Role__Authority", role)
			}
			c.AndOr(cond)

			return c.GetAny()
		}).AttemptThen(func(b bool) *result.Result[bool] {
			if !b {
				return result.OfError[bool](fault.NewUserUnauthorizedError("wrong role"))
			}
			return result.OfValue(true)
		})

		if len(auth.Data.TenantUUID) > 0 {
			return rio.MapToValue(roleChecker.AndThen(tenantChecker), auth)
		}
		return rio.MapToValue(roleChecker, auth)
	}
}
