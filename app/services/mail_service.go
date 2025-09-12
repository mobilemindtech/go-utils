package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/mobilemindtech/go-utils/app/models"
	"io/ioutil"
	"net/http"
	"strings"
)

type MailService struct {
	Controller                  beego.Controller
	PasswordRecoverTemplateName string

	ApiKey               string
	ApiAppName           string
	AppName              string
	AppUrl               string
	EmailDefault         string
	EmailPasswordDefault string

	MailServerUrl string
}

func NewMailService(data map[string]string) *MailService {
	c := &MailService{}
	c.ApiKey = data["apiKey"]
	c.ApiAppName = data["apiAppName"]
	c.EmailDefault = data["emailDefault"]
	c.EmailPasswordDefault = data["emailPasswordPadrao"]
	c.AppName = data["appName"]
	c.AppUrl = data["appUrl"]
	c.MailServerUrl = data["mailServerUrl"]
	return c
}

func (this *MailService) Send(email *models.Email) error {
	data := this.GetDefaultEmail()
	data["subject"] = email.Subject
	data["to"] = email.To
	data["cco"] = email.Cco
	data["body"] = email.Body
	return this.PostEmail(data)
}

func (this *MailService) SendPasswordRecover(to string, name string, token string) error {
	url := fmt.Sprintf("%v/password/change?token=%v", this.AppUrl, token)
	return this.SendPasswordRecoverWithUrl(to, name, url)
}

func (this *MailService) SendPasswordRecoverWithUrl(to string, name string, url string) error {

	this.Controller.TplName = this.PasswordRecoverTemplateName

	this.Controller.Data["user_name"] = name
	this.Controller.Data["recover_url"] = url

	content, err := this.Controller.RenderString()

	if err != nil {
		logs.Debug("error Controller.RenderString ", err.Error())
		return err
	}

	email := this.GetDefaultEmail()
	email["subject"] = fmt.Sprintf("%v - Recuperação de Senha", this.AppName)
	email["to"] = to
	email["body"] = content

	return this.PostEmail(email)
}

func (this *MailService) SendPasswordRecoverCode(to string, name string, code string) error {

	this.Controller.TplName = this.PasswordRecoverTemplateName

	this.Controller.Data["user_name"] = name
	this.Controller.Data["code"] = code

	content, err := this.Controller.RenderString()

	if err != nil {
		logs.Debug("error Controller.RenderString ", err.Error())
		return err
	}

	email := this.GetDefaultEmail()
	email["subject"] = fmt.Sprintf("%v - Recuperação de Senha", this.AppName)
	email["to"] = to
	email["body"] = content

	return this.PostEmail(email)
}

func (this *MailService) PostEmail(email map[string]string) error {

	if len(strings.TrimSpace(this.EmailDefault)) > 0 && len(strings.TrimSpace(this.EmailPasswordDefault)) > 0 {
		email["username"] = this.EmailDefault
		email["from"] = this.EmailDefault
		email["password"] = this.EmailPasswordDefault
	}

	email["application"] = this.ApiAppName

	jsonData, err := json.Marshal(email)

	if err != nil {
		logs.Debug("error json.Marshal ", err.Error())
		return err
	}

	signatureHash := this.GenerateHash(jsonData)

	data := bytes.NewBuffer(jsonData)

	logs.Debug("MAIL SERVER URL %v TO %v", this.MailServerUrl, email["to"])

	client := &http.Client{}
	req, err := http.NewRequest("POST", this.MailServerUrl, data)

	if err != nil {
		logs.Debug("error http.NewRequest ", err.Error())
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Hub-Signature", signatureHash)

	r, err := client.Do(req)

	if err != nil {
		logs.Debug("error http.Do ", err.Error())
		return err
	}

	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logs.Debug("error ioutil.ReadAll ", err.Error())
		return err
	}

	logs.Debug(string(response))

	return nil
}

func (this *MailService) OnTemplate(templateName string, values map[string]string) (string, error) {

	content, err := this.GetHtmlTemplate(templateName)

	if err != nil {
		return "", err
	}

	return this.ReplaceTemplate(content, values), nil
}

func (this *MailService) GetHtmlTemplate(templateName string) (string, error) {

	content := ""

	buffer, err := ioutil.ReadFile(fmt.Sprintf("conf/extra/%v.html", templateName))

	if err != nil {
		return content, err
	}

	return string(buffer), nil
}

func (this *MailService) ReplaceTemplate(templateContent string, values map[string]string) string {

	for k, v := range values {
		templateContent = strings.Replace(templateContent, k, v, -1)
	}

	return templateContent

}

func (this *MailService) GetDefaultEmail() map[string]string {

	emailMap := map[string]string{}

	emailMap["fromName"] = this.AppName
	emailMap["application"] = this.ApiAppName

	return emailMap
}

func (this *MailService) GenerateHash(body []byte) string {
	mac := hmac.New(sha1.New, []byte(this.ApiKey))

	bodyHex := []byte(hex.EncodeToString(body))

	mac.Write(bodyHex)
	rawBodyMAC := mac.Sum(nil)
	computedHash := base64.StdEncoding.EncodeToString(rawBodyMAC)
	return computedHash
}
