package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/beego/db"
	"io/ioutil"
	"net/http"
	"strings"
)

type Subscriber struct {
	Id    string
	Name  string
	Email string
	Dev   bool
}

type PushService struct {
	Session     *db.Session

	pushServerUser  string
	pushServerKey   string
	pushAppProdName string
	pushAppDevName  string
	pushServerUrl   string
	pushServerMode  string

	notificationTitle string
	notificationColor string
	notificationIcon  string
}

func NewPushService(session *db.Session) *PushService {
	pushServer := new(PushService)
	pushServer.Session = session

	return pushServer
}

func (this *PushService) Configure(data map[string]string) {

	this.pushAppProdName = data["AppProdName"]
	this.pushAppDevName = data["AppDevName"]
	this.pushServerUser = data["AccessUserName"]
	this.pushServerKey = data["AccessKey"]
	this.pushServerUrl = data["Server"]
	this.pushServerMode = data["Mode"]

	this.notificationTitle = data["NotificationTitle"]
	this.notificationColor = data["NotificationColor"]
	this.notificationIcon = data["NotificationIcon"]

}

func (this *PushService) NotificateUserName(username string, message string) error {
	list := []string{username}
	return this.NotificateByUserNameList(list, message)
}

// query should return a username list
func (this *PushService) NotificateByQuery(query string, message string) error {

	var results orm.ParamsList
	_, err := this.Session.GetDb().Raw(query).ValuesFlat(&results)

	if err != nil {
		logs.Debug("PushService.sendToAppUsers %v", err)
		return err
	}

	list := []string{}

	for _, username := range results {
		uname := username.(string)
		if len(strings.TrimSpace(uname)) > 0 {
			list = append(list, uname)
		}
	}

	if len(list) == 0 {
		logs.Debug("subscribers not found from query")
		return nil
	}

	return this.NotificateByUserNameList(list, message)

}

func (this *PushService) NotificateAll(message string) error {
	return this.NotificateByUserNameList(nil, message)
}

func (this *PushService) NotificateByUserNameList(usernameList []string, message string) error {

	if len(message) == 0 {
		return errors.New("notification message is empty")
	}

	notification := map[string]interface{}{}
	notification["msg"] = message
	notification["title"] = this.notificationTitle
	notification["icon"] = this.notificationIcon
	notification["color"] = this.notificationColor

	if usernameList == nil || len(usernameList) == 0 {
		return this.post("all",notification)
	} else {
		notification["users"] = usernameList
		return this.post("users",notification)
	}

	return nil
}

func (this *PushService) post(path string, notification map[string]interface{}) error {

	action := ""

	if this.pushServerMode == "PRODUCTION" {
		action = fmt.Sprintf("/event/%v/%v", this.pushAppProdName, path)
	} else {
		action = fmt.Sprintf("/event/%v/%v", this.pushAppDevName, path)
	}


	jsonData, err := json.Marshal(notification)

	if err != nil {
		logs.Debug("PushService.sendToAppUsers json.Marshal %v", err.Error())
		return err
	}

	data := bytes.NewBuffer(jsonData)

	client := &http.Client{}

	url := fmt.Sprintf("%v%v", this.pushServerUrl, action)

	logs.Debug("** send notification %v, channel %v", notification, url)

	req, err := http.NewRequest("POST", url, data)

	if err != nil {
		logs.Debug("PushService.FindUsers error http.NewRequest ", err.Error())
		return err
	}

	req.SetBasicAuth(this.pushServerUser, this.pushServerKey)
	req.Header.Set("Content-Type", "application/json")

	r, err := client.Do(req)

	if err != nil {
		logs.Debug("PushService.FindUsers error client.Do ", err.Error())
		return err
	}

	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		logs.Debug("PushService.sendToAppUsers ioutil.ReadAll %v", err.Error())
		return err
	}

	logs.Debug("PushService.sendToAppUsers post result %v", string(response))

	return nil
}
