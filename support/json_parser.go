package support

import (
  "github.com/astaxie/beego/context"
  "encoding/json"
  _"io/ioutil"
  "strconv"
  "time"
  "fmt"
  
)


type JsonParser struct {


}

func (c JsonParser) JsonToMap(ctx *context.Context) (map[string]interface{}, error) {
  data := make(map[string]interface{})
  err := json.Unmarshal(ctx.Input.RequestBody, &data) 
  return data, err 
}

func (c JsonParser) JsonToModel(ctx *context.Context, model interface{}) error { 
	fmt.Println("### %s", string(ctx.Input.RequestBody))
  err := json.Unmarshal(ctx.Input.RequestBody, &model)      
  return err
}

func (c JsonParser) GetJsonObject(json map[string]interface{}, key string) map[string]interface{} {
   
   if opt, ok := json[key]; ok {
    return opt.(map[string]interface{})
   }

   return nil  
}

func (c JsonParser) GetJsonInt(json map[string]interface{}, key string) int{
  val, _ := strconv.Atoi(json[key].(string))
  return val
}

func (c JsonParser) GetJsonInt64(json map[string]interface{}, key string) int64{
  val, _ := strconv.Atoi(json[key].(string))
  return int64(val)
}

func (c JsonParser) GetJsonString(json map[string]interface{}, key string) string{
  
  if json[key] == nil {
    return ""
  }

  val := json[key].(string)  

  if val == "null" || val == "undefined" {
    return ""
  }

  return val
}

func (c JsonParser) GetJsonDate(json map[string]interface{}, key string, layout string) time.Time{
  date, _ := time.Parse(layout, c.GetJsonString(json, key))
  return date
}
