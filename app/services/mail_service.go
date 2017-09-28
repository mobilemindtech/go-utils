package services

import (
	"github.com/mobilemindtec/go-utils/app/models"
  "github.com/astaxie/beego"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"bytes"
	"fmt"
)

type MailService struct {
	Controller beego.Controller
	PasswordRecoverTemplateName string
}

func NewMailService() *MailService{
	return &MailService{}
}

func (this *MailService) Send(email *models.Email) error{
	data := this.GetDefaultEmail()
	data["subject"] = email.Subject
	data["to"] = email.To
	data["cco"] = email.Cco
	data["body"] = email.Body
	return this.PostEmail(data)
}

func (this *MailService) SendPasswordRecover(to string, name string, token string) error {

	app_name := beego.AppConfig.String("app_name")
	app_url := beego.AppConfig.String("app_url")

  this.Controller.TplName = this.PasswordRecoverTemplateName

	this.Controller.Data["user_name"] = name
	this.Controller.Data["recover_url"] = fmt.Sprintf("%v/password/change?token=%v", app_url, token)

	content, err := this.Controller.RenderString()

	if err != nil {
		fmt.Println("error Controller.RenderString ", err.Error())
		return err
	}


	email := this.GetDefaultEmail()
	email["subject"] = fmt.Sprintf("%v - Recuperação de Senha", app_name)
	email["to"] = to
	email["body"] = content

	return this.PostEmail(email)
}

func (this *MailService) PostEmail(email map[string]string) error {

	mail_server_url := beego.AppConfig.String("mail_server_url")

	jsonData, err := json.Marshal(email)

	if err != nil {
		fmt.Println("error json.Marshal ", err.Error())
		return err
	}

	data := bytes.NewBuffer(jsonData)

	r, err := http.Post(mail_server_url, "text/json", data)

	if err != nil {
		fmt.Println("error http.Post ", err.Error())
		return err
	}

	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println("error ioutil.ReadAll ", err.Error())
		return err
	}

	fmt.Println(string(response))

	return nil
}


func (this *MailService) OnTemplate(templateName string, values map[string]string) (string, error){

	content, err := this.GetHtmlTemplate(templateName)

	if err != nil {
		return "", err
	}

	return this.ReplaceTemplate(content, values), nil
}

func (this *MailService) GetHtmlTemplate(templateName string) (string, error){

	content := ""

	buffer, err := ioutil.ReadFile(fmt.Sprintf("conf/extra/%v.html", templateName))

	if err != nil {
		return content, err
	}

	return string(buffer), nil
}

func (this *MailService) ReplaceTemplate(templateContent string, values map[string]string) string{

	for k, v := range values {
		templateContent = strings.Replace(templateContent, k, v, -1)
	}

	return templateContent

}

func (this *MailService) GetDefaultEmail() map[string]string {
	email := beego.AppConfig.String("email")
	//email_username := beego.AppConfig.String("email_username")
	email_password := beego.AppConfig.String("email_password")
	email_from := beego.AppConfig.String("email_from")

	emailMap := map[string]string{}

	emailMap["fromName"] = email_from
	emailMap["application"] = "goapp"

	emailMap["gmailUserName"] = email
	emailMap["gmailPassword"] = email_password

	return emailMap
}
