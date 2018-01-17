package services

import (
  "github.com/mobilemindtec/go-utils/app/models"
	"github.com/mobilemindtec/go-utils/beego/db"
  "github.com/mobilemindtec/go-utils/app/util"
  "github.com/astaxie/beego"  
  "github.com/beego/i18n"
  "errors"
  "time"
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

		return user, errors.New(this.GetMessage("login.invalidToken"))

	} else if !user.Enabled {

		return user, errors.New(this.GetMessage("login.inactiveMsg"))

	}else if time.Now().In(util.GetDefaultLocation()).Unix() > user.ExpirationDate.Unix() {

		return user, errors.New(this.GetMessage("login.expiredToken"))

	}else if !byToken && !user.IsSamePassword(password) {
		beego.Debug("### password not match ")
		// No matched password
		return user, errors.New(this.GetMessage("login.invalid"))

	}else {

		tenant, err := this.ModelTenantUser.GetFirstTenant(user)

		if err != nil {
			return user, errors.New(this.GetMessage("login.error"))
		}

		if tenant == nil {
			return user, errors.New("user does not has active tenant related")	
		}		

		user.LastLogin = time.Now().In(util.GetDefaultLocation())
		if err := this.Session.Update(user); err != nil {
			beego.Debug("### update user login error %v", err)
		}
		return user, nil

	}	
}


func (this *LoginService) GetMessage(key string, args ...interface{}) string{
  return i18n.Tr(this.Lang, key, args)
}
