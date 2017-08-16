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
}

func NewLoginService(lang string, session *db.Session) *LoginService {
	return &LoginService{ Lang: lang, Session: session }
}

func (this *LoginService) Authenticate(username string, password string) (user *models.User, err error)  {

	user = models.NewUser(this.Session)

	user, err = user.GetByUserName(username)

	if err != nil {

		if err.Error() == "<QuerySeter> no row found" {
			err = errors.New(this.GetMessage("login.invalid"))
		}

		return user, err

	} else if user == nil || user.Id < 1 {

		return user, errors.New(this.GetMessage("login.invalid"))

	} else if !user.IsSamePassword(password) {

		return user, errors.New(this.GetMessage("login.invalid"))

	} else if !user.Enabled {

		return user, errors.New(this.GetMessage("login.inactiveMsg"))

	}else {

		user.LastLogin = time.Now().In(util.GetDefaultLocation())
		if err := this.Session.Update(user); err != nil {
			beego.Debug("### update user login error %v", err)
		}
		return user, nil

	}
}


func (this *LoginService) AuthenticateToken(token string) (*models.User, error) {
	user := models.NewUser(this.Session)
	var err error


	user, err = user.GetByToken(token)

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

	}else {

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
