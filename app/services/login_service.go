package services

import (
  "github.com/mobilemindtec/go-utils/app/models"
	"github.com/mobilemindtec/go-utils/beego/db"
	"github.com/beego/beego/v2/core/logs"
  _ "github.com/mobilemindtec/go-utils/app/util"  
  "github.com/beego/i18n"
  "errors"
  _ "time"  
)

type LoginService struct {
	Lang string
	Session *db.Session
	ModelUser *models.User
	ModelTenantUser *models.TenantUser
}

func NewLoginService(lang string, session *db.Session) *LoginService {
	return &LoginService{ Lang: lang, Session: session, ModelUser: models.NewUser(session), ModelTenantUser: models.NewTenantUser(session) }
}

func (this *LoginService) Authenticate(username string, password string) (*models.User, error)  {
	user, err := this.ModelUser.GetByUserName(username)
	return this.Login(user, password, false, err)
}


func (this *LoginService) AuthenticateToken(token string) (*models.User, error) {
	user, err := this.ModelUser.GetByToken(token)
	return this.Login(user, "", true, err)
}

func (this *LoginService) Login(user *models.User, password string, byToken bool, err error) (*models.User, error){
	if err != nil {

		if err.Error() == "<QuerySeter> no row found" {
			err = errors.New(this.GetMessage("login.invalid"))
		}

		return user, err

	} else if user == nil || user.Id < 1 {

		logs.Debug("### user not found %v", user)
		return user, errors.New(this.GetMessage("login.invalid"))

	} else if !user.Enabled {

		logs.Debug("### user not enabled ")
		return user, errors.New(this.GetMessage("login.inactiveMsg"))

	}else if !byToken && !user.IsSamePassword(password) {
		logs.Debug("### password not match ")
		// No matched password
		return user, errors.New(this.GetMessage("login.invalid"))

	}else {

		tenant, err := this.ModelTenantUser.GetFirstTenant(user)

		if err != nil {
			logs.Debug("### error on get user tenant %v", err)
			return user, errors.New(this.GetMessage("login."))
		}

		if tenant == nil {
			logs.Debug("### error does not have tenant")
			return user, errors.New("user does not has active tenant related")	
		}		


		models.NewUser(this.Session).UpdateLastLogin(user.Id)


		return user, nil

	}	
}


func (this *LoginService) GetMessage(key string, args ...interface{}) string{
  return i18n.Tr(this.Lang, key, args)
}
