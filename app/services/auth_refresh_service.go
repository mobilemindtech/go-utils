package services

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-io/option"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-io/rio"
	"github.com/mobilemindtech/go-io/types/unit"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/fault"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	uuid "github.com/satori/go.uuid"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
)

type RefreshData struct {
	RefreshToken string `json:"refresh_token" valid:"Required"`
}

// AuthClient autentica um cliente de loja (site/app) e emite bearer + refresh token.
// tenantUuid é a loja do header X-Auth-Tenant; o cliente precisa estar vinculado a ela.
func (this *AuthService) AuthClient(body []byte, tenantUuid string, userAgent string) *rio.IO[*AuthResult] {

	withTenant := rio.AttemptThen(this.authenticate(body), func(auth *AuthUser) *result.Result[*AuthUser] {
		if len(strings.TrimSpace(tenantUuid)) > 0 {
			auth.Data.TenantUUID = tenantUuid
		}
		if len(strings.TrimSpace(auth.Data.TenantUUID)) == 0 {
			return result.OfError[*AuthUser](fault.NewUserUnauthorizedError("tenant required"))
		}
		return result.OfValue(auth)
	})

	return rio.AttemptThen(
		rio.AttemptThenOfIO(withTenant, this.checkPermissions([]string{"ROLE_CLIENTE"}, false)),
		func(auth *AuthUser) *result.Result[*AuthResult] {
			return this.IssueTokens(auth.User, "", userAgent)
		})
}

// IssueTokens emite bearer + refresh token para o usuário (cliente ou root/admin). family vazio inicia
// uma nova família (novo login); no refresh a família do token anterior é mantida.
func (this *AuthService) IssueTokens(user *models.User, family string, userAgent string) *result.Result[*AuthResult] {

	if err := this.ensureUserUuid(user); err != nil {
		return result.OfError[*AuthResult](err)
	}

	access := this.newAccessToken(user, AccessTokenTTL)

	return result.FlatMap(access, func(auth *AuthResult) *result.Result[*AuthResult] {

		raw, err := newOpaqueToken()
		if err != nil {
			return result.OfError[*AuthResult](err)
		}

		if len(family) == 0 {
			family = uuid.NewV4().String()
		}

		if len(userAgent) > 255 {
			userAgent = userAgent[:255]
		}

		entity := &models.UserRefreshToken{
			TokenHash: support.TextToSha256Hex(raw),
			Family:    family,
			ExpiresAt: util.DateNow().Add(RefreshTokenTTL),
			UserAgent: userAgent,
			User:      user,
		}

		if err := this.Session.Save(entity); err != nil {
			return result.OfError[*AuthResult](err)
		}

		auth.RefreshToken = raw
		auth.RefreshExpiresAt = entity.ExpiresAt.Unix()
		return result.OfValue(auth)
	})
}

// Refresh troca um refresh token válido por um novo par bearer + refresh (rotação).
// O token usado é revogado; reutilizar um token já revogado revoga a família inteira.
// Qualquer falha retorna 401, o cliente deve então fazer login novamente.
func (this *AuthService) Refresh(body []byte, userAgent string) *rio.IO[*AuthResult] {

	return rio.AttemptThen(this.findRefreshToken(body), func(opt *option.Option[*models.UserRefreshToken]) *result.Result[*AuthResult] {

		if opt.IsNone() {
			return invalidRefreshToken()
		}

		token := opt.Get()
		now := util.DateNow()

		if token.IsRevoked() {
			logs.Warning("refresh token reuse detected, user %v, family %v", token.User.Id, token.Family)
			revokeFamilyNoTx(token.Family)
			return invalidRefreshToken()
		}

		if token.IsExpired(now) || token.User == nil || !token.User.Enabled {
			return invalidRefreshToken()
		}

		// revogação condicional: duas requisições concorrentes com o mesmo token
		// não podem ambas rotacionar, a segunda é tratada como reuso
		affected, err := db.NewRawQueryArgs(this.Session,
			"update user_refresh_tokens set revoked_at = ? where id = ? and revoked_at is null",
			now, token.Id).Execute()

		if err != nil {
			return result.OfError[*AuthResult](err)
		}

		if affected == 0 {
			logs.Warning("refresh token concurrent reuse, user %v, family %v", token.User.Id, token.Family)
			revokeFamilyNoTx(token.Family)
			return invalidRefreshToken()
		}

		return this.IssueTokens(token.User, token.Family, userAgent)
	})
}

// RevokeRefresh logout: revoga a família do refresh token informado (a sessão
// daquele dispositivo). Token inexistente não é erro.
func (this *AuthService) RevokeRefresh(body []byte) *rio.IO[*unit.Unit] {
	return rio.AttemptThen(this.findRefreshToken(body), func(opt *option.Option[*models.UserRefreshToken]) *result.Result[*unit.Unit] {
		if opt.IsSome() {
			if err := revokeFamily(this.Session, opt.Get().Family); err != nil {
				return result.OfError[*unit.Unit](err)
			}
		}
		return result.OfValue(unit.OfUnit())
	})
}

// RevokeAllRefreshTokens revoga todas as sessões do usuário (ex.: troca de senha).
func (this *AuthService) RevokeAllRefreshTokens(user *models.User) error {
	_, err := db.NewRawQueryArgs(this.Session,
		"update user_refresh_tokens set revoked_at = ? where user_id = ? and revoked_at is null",
		util.DateNow(), user.Id).Execute()
	return err
}

// DeleteOldRefreshTokens remove refresh tokens expirados ou revogados há mais de
// retention. Revogados são mantidos por um tempo para a detecção de reuso.
func DeleteOldRefreshTokens(session *db.Session, retention time.Duration) (int64, error) {
	limit := util.DateNow().Add(-retention)
	return db.NewRawQueryArgs(session,
		"delete from user_refresh_tokens where expires_at < ? or revoked_at < ?",
		limit, limit).Execute()
}

// ensureUserUuid usuários antigos não têm uuid, que é o identificador do bearer token
func (this *AuthService) ensureUserUuid(user *models.User) error {
	if len(strings.TrimSpace(user.Uuid)) > 0 {
		return nil
	}
	user.Uuid = models.NewUser(this.Session).GenereteUuid()
	_, err := db.NewRawQueryArgs(this.Session,
		"update users set uuid = ? where id = ?", user.Uuid, user.Id).Execute()
	return err
}

func (this *AuthService) findRefreshToken(body []byte) *rio.IO[*option.Option[*models.UserRefreshToken]] {

	jsonParser := rio.Attempt(func() *result.Result[*RefreshData] {
		return json.UnmarshalResult[*RefreshData](body)
	})

	entityValidator := rio.AttemptThen(jsonParser, func(data *RefreshData) *result.Result[*RefreshData] {
		res := validator.New().
			WithPath("RefreshData").
			ValidateOrError(data)
		return result.MapToValue(res, data)
	})

	return rio.AttemptThen(entityValidator, func(data *RefreshData) *result.Result[*option.Option[*models.UserRefreshToken]] {
		return criteria.New[*models.UserRefreshToken](this.Session).
			Eq("TokenHash", support.TextToSha256Hex(data.RefreshToken)).
			Eager("User").
			GetFirst()
	})
}

func revokeFamily(session *db.Session, family string) error {
	_, err := db.NewRawQueryArgs(session,
		"update user_refresh_tokens set revoked_at = ? where family = ? and revoked_at is null",
		util.DateNow(), family).Execute()
	return err
}

// revokeFamilyNoTx revoga fora da transação da requisição, que sofre rollback
// porque a requisição termina em erro (401)
func revokeFamilyNoTx(family string) {
	err := db.RunNoTx(nil, func(session *db.Session) error {
		return revokeFamily(session, family)
	})
	if err != nil {
		logs.Error("error to revoke refresh token family %v: %v", family, err)
	}
}

func invalidRefreshToken() *result.Result[*AuthResult] {
	return result.OfError[*AuthResult](fault.NewUserNotFoundError("invalid refresh token"))
}

func newOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
