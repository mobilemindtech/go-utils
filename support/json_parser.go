package support

import (
  "github.com/mobilemindtec/go-utils/app/util"
  "github.com/astaxie/beego/context"
  "encoding/json"
  "strconv"
  "strings"
  "errors"
  "time"
  "fmt"

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

  if err != nil {
    return errors.New(fmt.Sprintf("error on JsonToModel.json.Unmarshal: %v", err.Error()))
  }

  return nil
}


func (c JsonParser) FormToJson(ctx *context.Context) map[string]interface{} {
  return c.FormToJsonWithFieldsConfigs(ctx, nil);
}

func (c JsonParser) FormToJsonWithFieldsConfigs(ctx *context.Context, configs map[string]string) map[string]interface{} {

  jsonMap := make(map[string]interface{})

  data := ctx.Request.Form

  for k, v := range  data{

    //this.Log("key %v, value = %v", k, v)

    if len(v) == 0 {
      continue
    }

    if strings.Contains(k, ".") {
      keys := strings.Split(k, ".")

      parent := jsonMap

      for i, key := range keys {

        if _, ok := parent[key].(map[string]interface{}); !ok {
          parent[key] = make(map[string]interface{})
        }

        if i < len(keys) -1 {
          parent = parent[key].(map[string]interface{})
        } else {
          parent[key] = v[0]

          if configs != nil {
            for field, format := range configs {
              if field == key {
                if parent[key] != nil {
                  auxDate, _ := util.DateParse(format, parent[key].(string))
                  jsonDateLayout := "2006-01-02T15:04:05-07:00"
                  parent[key] = auxDate.Format(jsonDateLayout)
                }
              }
            }
          }

        }
      }

    } else {
      jsonMap[k] = v[0]
      if configs != nil {
        for field, format := range configs {
          if field == k {
            if jsonMap[k] != nil {
              auxDate, _ := util.DateParse(format, jsonMap[k].(string))
              jsonDateLayout := "2006-01-02T15:04:05-07:00"
              jsonMap[k] = auxDate.Format(jsonDateLayout)
            }
          }
        }
      }
    }
  }

  return jsonMap
}

func (c JsonParser) FormToModel(ctx *context.Context, model interface{}) error {
  return c.FormToModelWithFieldsConfigs(ctx, model, nil)
}

func (c JsonParser) FormToModelWithFieldsConfigs(ctx *context.Context, model interface{}, configs map[string]string) error {

  jsonMap := c.FormToJsonWithFieldsConfigs(ctx, configs)

  jsonData, err := json.Marshal(jsonMap)

  if err != nil {
    return errors.New(fmt.Sprintf("error on FormToModel.json.Marshal: %v", err.Error()))
  }

  err = json.Unmarshal(jsonData, model)

  if err != nil {
    return errors.New(fmt.Sprintf("error on FormToModel.json.Unmarshal: %v", err.Error()))
  }

  return nil

}

func (c JsonParser) GetJsonObject(json map[string]interface{}, key string) map[string]interface{} {

   if c.HasJsonKey(json, key) {
    opt, _ := json[key]
    if opt != nil {
      return opt.(map[string]interface{})
    }
   }

   return nil
}

func (c JsonParser) GetJsonArray(json map[string]interface{}, key string) []map[string]interface{} {

   if c.HasJsonKey(json, key) {
    opt, _ := json[key]

    items := new([]map[string]interface{})

    if array, ok := opt.([]interface{}); ok {
      for _, it := range array {
        if p, ok := it.(map[string]interface{}); ok {
          *items = append(*items, p)
        }
      }
    }

    return *items
   }

   return nil
}

func (c JsonParser) GetJsonSimpleArray(json map[string]interface{}, key string) []interface{} {

   if c.HasJsonKey(json, key) {
    opt, _ := json[key]

    if array, ok := opt.([]interface{}); ok {
      return array
    }

   }

   return nil
}

func (c JsonParser) GetArrayFromJson(json map[string]interface{}, key string) []interface{} {

   if c.HasJsonKey(json, key) {
    opt, _ := json[key]

    items := new([]interface{})

    if array, ok := opt.([]interface{}); ok {
      for _, it := range array {
        //if p, ok := it.(map[string]interface{}); ok {
          *items = append(*items, it)
        //}
      }
    }

    return *items
   }

   return nil
}

func (c JsonParser) GetJsonInt(json map[string]interface{}, key string) int{
  var val int

  if c.HasJsonKey(json, key) {
    if _, ok := json[key].(int); ok {
      val = json[key].(int)
    } else if _, ok := json[key].(int64); ok {
      val = int(json[key].(int64))
    } else if _, ok := json[key].(float64); ok {
      val = int(json[key].(float64))
    } else {
      val, _ = strconv.Atoi(c.GetJsonString(json, key))
    }
  } else {
    fmt.Println("not has int key %v", key)
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
    } else if _, ok := json[key].(float64); ok {
      val = int(json[key].(float64))
    } else {
      val, _ = strconv.Atoi(c.GetJsonString(json, key))
    }
  } else {
    fmt.Println("not has int key %v", key)
  }

  return int64(val)
}

func (c JsonParser) GetJsonBool(json map[string]interface{}, key string) bool{

  var val bool

  if c.HasJsonKey(json, key) {
    if _, ok := json[key].(bool); ok {
      val = json[key].(bool)
    } else {
      val, _ = strconv.ParseBool(c.GetJsonString(json, key))
    }
  }

  return val
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

func (c JsonParser) JsonInterfaceToInt64(item interface{}) int64{

  var val int = 0

  if _, ok := item.(int); ok {
    val = item.(int)
  } else if _, ok := item.(int64); ok {
    val = int(item.(int64))
  } else if _, ok := item.(float64); ok {
    val = int(item.(float64))
  } else {
    val, _ = strconv.Atoi(fmt.Sprintf("%v", item))
  }

  return int64(val)
}

func (c JsonParser) GetJsonDate(json map[string]interface{}, key string, layout string) time.Time{
  date, _ := time.ParseInLocation(layout, c.GetJsonString(json, key), c.DefaultLocation)
  return date
}


func (c JsonParser) HasJsonKey(json map[string]interface{}, key string) bool{
  if _, ok := json[key]; ok {
    return true
  }
  return false
}
