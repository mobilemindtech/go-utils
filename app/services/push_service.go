package services

import (
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/astaxie/beego/orm"
	"encoding/json"
  "io/ioutil"
  "net/http"
  "bytes"
  "fmt"
)

type Subscriber struct {
	Id string
	Name string
	Email string
	Dev bool
}

type PushService struct{

	Session *db.Session
	Subscribers map[string][]*Subscriber

	pushServerUser string
	pushServerKey string
	pushAppProdName string
	pushAppDevName string
	pushServerUrl string
	pushServerMode string

	notificationTitle string
	notificationColor string
	notificationIcon string

}

func NewPushService(session *db.Session) *PushService{
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

func (this *PushService) GetSubscribersFromPushServer(url string) (map[string]interface{}, error){
	

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println("PushService.FindUsers error http.NewRequest ", err.Error())
		return nil, err
	}
	req.SetBasicAuth(this.pushServerUser, this.pushServerKey)

	r, err := client.Do(req)

	if err != nil {
		fmt.Println("PushService.FindUsers error client.Do ", err.Error())
		return nil, err
	}

  jsonData := make(map[string]interface{})

  body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Println("PushService.FindUsers error ioutil.ReadAll ", err.Error())
		return nil, err
	}

  err = json.Unmarshal(body, &jsonData)

	if err != nil {
		fmt.Println("PushService.FindUsers error json.Unmarshal ", err.Error())
		return nil, err
	}

	//fmt.Println("data=%v", jsonData)

	return jsonData, nil

}

func (this *PushService) LoadSubscribers() error{

	this.Subscribers = map[string][]*Subscriber{}

  fmt.Println("PushService.LoadSubscribers run")


  process := func(jsonData map[string]interface{}, dev bool){
		for key, value := range jsonData {

			// key = username
			// value = list of subscriptions

			//fmt.Println("key=%v", key)

			if _, ok := this.Subscribers[key]; !ok {
				this.Subscribers[key] = []*Subscriber{}
			}

			//fmt.Println("value=%v", value)

			jsonSubscribers := 	value.([]interface{})

			for _, sub := range jsonSubscribers {

				subJson := sub.(map[string]interface{})
				name := subJson["name"].(string)
				email := subJson["email"].(string)
				subscrible_id := subJson["subscrible_id"].(string)

				subscriber := Subscriber{ Name: name, Email: email, Id: subscrible_id, Dev: dev }

				//fmt.Println("subscriber=%v", subscriber)

				this.Subscribers[key] = append(this.Subscribers[key], &subscriber)
			}

		}  	
  }


	if this.pushServerMode == "PRODUCTION" || this.pushServerMode == "ALL" {
		action := fmt.Sprintf("/apps/%v", this.pushAppProdName)

		data, err := this.GetSubscribersFromPushServer(fmt.Sprintf("%v%v", this.pushServerUrl, action))

		if err != nil {
			return err
		}

		process(data, false)
	}

	if this.pushServerMode == "TEST" || this.pushServerMode == "ALL" {
		action := fmt.Sprintf("/apps/%v", this.pushAppDevName)

		data, err := this.GetSubscribersFromPushServer(fmt.Sprintf("%v%v", this.pushServerUrl, action))

		if err != nil {
			return err
		}

		process(data, true)
	}


	fmt.Println("subscribers=%v", this.Subscribers)

	return nil
}




func (this *PushService) NotificateUserName(username string, message string) error{
  list := []string{username}
 	return  this.NotificateByUserNameList(&list, message)
}

//
// query should return a username list
//
func (this *PushService) NotificateByQuery(query string, message string) error{

  var results orm.ParamsList
	_, err := this.Session.Db.Raw(query).ValuesFlat(&results)

  if err != nil {
		fmt.Println("PushService.sendToAppUsers %v", err)
		return err
	}

  list := []string{}

  for _, username := range results {
    list = append(list, username.(string))
  }

  if len(list) == 0 {
  	fmt.Println("not found subscribers from query")
  	return nil
  }

  return this.NotificateByUserNameList(&list, message)


}

func (this *PushService) NotificateAll(message string) error{
	return this.NotificateByUserNameList(nil, message)	
}

func (this *PushService) NotificateByUserNameList(usernameList *[]string, message string) error{
	

	notification := map[string]string{}
	notification["msg"] = message
	notification["title"] = this.notificationTitle
	notification["icon"] = this.notificationIcon
	notification["color"] = this.notificationColor

	if usernameList == nil || len(*usernameList) == 0 {
		return this.post(notification, nil)
	} else {

		this.LoadSubscribers()

		for _, username := range *usernameList {

	    //fmt.Println("this.Subscribers=%v", this.Subscribers)

			subscribers, ok := this.Subscribers[username]

			if !ok {

				fmt.Println("subscriber %v not found at push server %v %v", username, ok, subscribers)

				continue
			}


			for _, subscriber := range subscribers {
				notification["data.user_id"] = subscriber.Id
				if err := this.post(notification, subscriber); err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func (this *PushService) post(notification map[string]string, subscriber *Subscriber) error {

		action := ""


		if subscriber == nil {

			if this.pushServerMode == "PRODUCTION" {
				action = fmt.Sprintf("/event/%v", this.pushAppProdName)
			} else {
				action = fmt.Sprintf("/event/%v", this.pushAppDevName)
			}

		} else {
			if subscriber.Dev {
				action = fmt.Sprintf("/event/%v", this.pushAppDevName)
			}	else {
				action = fmt.Sprintf("/event/%v", this.pushAppProdName)
			}
		}

		jsonData, err := json.Marshal(notification)

		if err != nil {
			fmt.Println("PushService.sendToAppUsers json.Marshal %v", err.Error())
			return err
		}


		data := bytes.NewBuffer(jsonData)

		client := &http.Client{}		

		url := fmt.Sprintf("%v%v", this.pushServerUrl, action)
		fmt.Println("send notification %v to %v", notification, url)

		req, err := http.NewRequest("POST", url, data)

		if err != nil {
			fmt.Println("PushService.FindUsers error http.NewRequest ", err.Error())
			return err
		}

		req.SetBasicAuth(this.pushServerUser, this.pushServerKey)
		req.Header.Set("Content-Type", "application/json")

		r, err := client.Do(req)

		if err != nil {
			fmt.Println("PushService.FindUsers error client.Do ", err.Error())
			return err
		}

		response, err := ioutil.ReadAll(r.Body)

		if err != nil {
			fmt.Println("PushService.sendToAppUsers ioutil.ReadAll %v", err.Error())
			return err
		}

		fmt.Println("PushService.sendToAppUsers post result %v", string(response))	

		return nil
}
