package support

import (
  "github.com/astaxie/beego/context"
  "encoding/json"  
  "strconv"
  "time"
  _"fmt"
  
)


type JsonParser struct {

  DefaultLocation *time.Location

}

func (c JsonParser) JsonToMap(ctx *context.Context) (map[string]interface{}, error) {
  data := make(map[string]interface{})
  err := json.Unmarshal(ctx.Input.RequestBody, &data) 
  return data, err 
}

func (c JsonParser) JsonToModel(ctx *context.Context, model interface{}) error { 
	//fmt.Println("### %s", string(ctx.Input.RequestBody))
  err := json.Unmarshal(ctx.Input.RequestBody, &model)      
  return err
}

func (c JsonParser) GetJsonObject(json map[string]interface{}, key string) map[string]interface{} {
   
   if c.HasJsonKey(json, key) {
    opt, _ := json[key]
    return opt.(map[string]interface{})
   }

   return nil  
}

func (c JsonParser) GetJsonInt(json map[string]interface{}, key string) int{
  var val int 

  if c.HasJsonKey(json, key) {
    if _, ok := json[key].(int); ok {
      val = json[key].(int)
    } else {
      val, _ = strconv.Atoi(c.GetJsonString(json, key))
    }
  }

  return val
}

func (c JsonParser) GetJsonInt64(json map[string]interface{}, key string) int64{

  var val int 

  if c.HasJsonKey(json, key) {
    if _, ok := json[key].(int); ok {
      val = json[key].(int)
    } else if _, ok := json[key].(int64); ok {
      val = int(json[key].(int64))
    } else {
      val, _ = strconv.Atoi(c.GetJsonString(json, key))
    }
  }

  return int64(val)
}

func (c JsonParser) GetJsonString(json map[string]interface{}, key string) string{
  
  var val string

  if !c.HasJsonKey(json, key) {
    return val
  }

  if _, ok := json[key].(string); ok {  

    val = json[key].(string)

    if val == "null" || val == "undefined" {
      return val
    }
  }

  return val
}

func (c JsonParser) GetJsonDate(json map[string]interface{}, key string, layout string) time.Time{
  date, _ := time.ParseInLocation(layout, c.GetJsonString(json, key), c.DefaultLocation)
  return date
}


func (C JsonParser) HasJsonKey(json map[string]interface{}, key string) bool{
  if _, ok := json[key]; ok {
    return true
  }
  return false
}