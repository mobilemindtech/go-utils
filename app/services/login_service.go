package services

import (
	"errors"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/i18n"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/beego/db"
	_ "time"
)

type LoginService struct {
	Lang            string
	Session         *db.Session
	ModelUser       *models.User
	ModelTenantUser *models.TenantUser
}

func NewLoginService(lang string, session *db.Session) *LoginService {
	return &LoginService{Lang: lang, Session: session, ModelUser: models.NewUser(session), ModelTenantUser: models.NewTenantUser(session)}
}

func (this *LoginService) Authenticate(username string, password string) (*models.User, error) {
	user, err := this.ModelUser.GetByUserName(username)
	return this.Login(user, password, false, err)
}

func (this *LoginService) AuthenticateToken(token string) (*models.User, error) {
	user, err := this.ModelUser.GetByToken(token)
	return this.Login(user, "", true, err)
}

func (this *LoginService) Login(user *models.User, password string, byToken bool, err error) (*models.User, error) {
	if err != nil {

		if err.Error() == "<QuerySeter> no row found" {
			err = LoginUserNotFound(this.GetMessage("login.invalid"))
		}

		return user, err

	} else if user == nil || user.Id < 1 {

		logs.Error("### user not found %v", user)
		return user, LoginUserNotFound(this.GetMessage("login.invalid"))

	} else if !user.Enabled {

		logs.Error("### user not enabled ")
		return user, LoginUserInactive(this.GetMessage("login.inactiveMsg"))

	} else if !byToken && !user.IsSamePassword(password) {
		logs.Error("### password not match ")
		// No matched password
		return user, LoginWrongPassword(this.GetMessage("login.invalid"))

	} else {

		tenant, err := this.ModelTenantUser.GetFirstTenant(user)

		if err != nil {
			logs.Error("### error on get user tenant %v", err)
			return user, errors.New(this.GetMessage("login.error", err.Error()))
		}

		if tenant == nil {
			logs.Error("### error does not have tenant")
			return user, LoginTenantNotFound(this.GetMessage("login.tenantNotFound"))
		}

		models.NewUser(this.Session).UpdateLastLogin(user.Id)

		return user, nil

	}
}

func (this *LoginService) GetMessage(key string, args ...interface{}) string {
	return i18n.Tr(this.Lang, key, args)
}
