package services

import (
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/astaxie/beego/orm"
  "github.com/astaxie/beego"
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

}

func NewPushService(session *db.Session, appName string) *PushService{
	pushServer := new(PushService)
  pushServer.Session = session


	pushServer.build()
	return pushServer
}

func (this *PushService) build() {
  this.pushAppProdName = beego.AppConfig.String("push_app_prod_name")
  this.pushAppDevName = beego.AppConfig.String("push_app_dev_name")
  this.pushServerUser = beego.AppConfig.String("push_app_user_name")
  this.pushServerKey = beego.AppConfig.String("push_app_user_password")
  this.pushServerUrl = beego.AppConfig.String("push_server_url")
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

func (this *PushService) LoadSubscribers() {

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


	
	action := fmt.Sprintf("/apps/%v", this.pushAppProdName)

	data, err := this.GetSubscribersFromPushServer(fmt.Sprintf("%v%v", this.pushServerUrl, action))

	if err != nil {
		return
	}

	process(data, false)

	action = fmt.Sprintf("/apps/%v", this.pushAppDevName)

	data, err = this.GetSubscribersFromPushServer(fmt.Sprintf("%v%v", this.pushServerUrl, action))

	if err != nil {
		return
	}

	process(data, true)


	fmt.Println("subscribers=%v", this.Subscribers)
}




func (this *PushService) NotificateUserName(username string, message string) {
  list := []string{username}
  this.NotificateByUserNameList(&list, message)
}

//
// query should return a username list
//
func (this *PushService) NotificateByQuery(query string, message string) {

  var results orm.ParamsList
	_, err := this.Session.Db.Raw(query).ValuesFlat(&results)

  if err != nil {
		fmt.Println("PushService.sendToAppUsers %v", err)
		return
	}

  list := []string{}

  for _, username := range results {
    list = append(list, username.(string))
  }

  this.NotificateByUserNameList(&list, message)
}

func (this *PushService) NotificateByUserNameList(usernameList *[]string, message string) {

	this.LoadSubscribers()

	
	action := fmt.Sprintf("/event/%v", this.pushAppProdName)


	for _, username := range *usernameList {

    //fmt.Println("this.Subscribers=%v", this.Subscribers)

		subscribers, ok := this.Subscribers[username]

		if !ok {

			fmt.Println("subscriber %v not found at push server %v %v", username, ok, subscribers)

			continue
		}


		for _, subscriber := range subscribers {

			notification := map[string]string{}
			notification["msg"] = message
			notification["title"] = "IMOSIG"
			notification["icon"] = "ic_stat_notify"
			notification["color"] = "#b20000"
			notification["data.user_id"] = subscriber.Id


			jsonData, err := json.Marshal(notification)

			if err != nil {
				fmt.Println("PushService.sendToAppUsers json.Marshal %v", err.Error())
				continue
			}

			fmt.Println("send notification %v", notification)

			data := bytes.NewBuffer(jsonData)

			client := &http.Client{}

			naction := action
			if subscriber.Dev {
				naction += "-dev"
			}

			req, err := http.NewRequest("POST", fmt.Sprintf("%v%v", this.pushServerUrl, naction), data)

			if err != nil {
				fmt.Println("PushService.FindUsers error http.NewRequest ", err.Error())
				return
			}
			req.SetBasicAuth(this.pushServerUser, this.pushServerKey)
			req.Header.Set("Content-Type", "application/json")

			r, err := client.Do(req)

			if err != nil {
				fmt.Println("PushService.FindUsers error client.Do ", err.Error())
				return
			}

			response, err := ioutil.ReadAll(r.Body)

			if err != nil {
				fmt.Println("PushService.sendToAppUsers ioutil.ReadAll %v", err.Error())
				continue
			}

			fmt.Println("PushService.sendToAppUsers post result %v", string(response))
		}

	}
}
