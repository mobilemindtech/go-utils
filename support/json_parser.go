package support

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/mobilemindtec/go-utils/app/util"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type FormatType int64

const (
	FormatTypeFloat FormatType = iota + 1
	FormatTypeInt
	FormatTypeDate
	FormatTypeBool
)

type FormJsonConfig struct {
	FieldName  string
	Parser     func(val string) interface{}
	FormatType FormatType
	Layout     string
}

func NewFormJsonConfig(fieldName string, formatType FormatType) *FormJsonConfig {
	return &FormJsonConfig{FieldName: fieldName, FormatType: formatType}
}

func (this *FormJsonConfig) SetLayout(layout string) *FormJsonConfig {
	this.Layout = layout
	return this
}

func (this *FormJsonConfig) SetParser(parser func(val string) interface{}) *FormJsonConfig {
	this.Parser = parser
	return this
}

func (this *FormJsonConfig) List() []*FormJsonConfig {
	return []*FormJsonConfig{this}
}

type JsonParser struct {
	DefaultLocation *time.Location
}

func NewJsonParser() *JsonParser {
	return &JsonParser{}
}

func ParseUrlValues[T any](data url.Values, modelOpts ...*T) *optional.Optional[*T] {

	var model T

	if len(modelOpts) > 0 {
		model = *modelOpts[0]
	}

	parser := NewJsonParser()
	values := parser.UrlValuesToMap(data)

	jsonData, err := json.Marshal(values)

	if err != nil {
		return optional.OfFail[*T](err)
	}

	err = json.Unmarshal(jsonData, &model)

	if err != nil {
		return optional.OfFail[*T](err)
	}

	return optional.Of[*T](&model)
}

func (this *JsonParser) JsonToMap(ctx *context.Context) (map[string]interface{}, error) {
	return this.JsonBytesToMap(ctx.Input.RequestBody)
}

func (this *JsonParser) JsonBytesToMap(body []byte) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(body, &data)
	return data, err
}

func (this *JsonParser) JsonToModel(ctx *context.Context, model interface{}) error {
	//fmt.Println("### %s", string(ctx.Input.RequestBody))
	err := json.Unmarshal(ctx.Input.RequestBody, &model)

	if err != nil {
		return errors.New(fmt.Sprintf("error on JsonToModel.json.Unmarshal: %v", err.Error()))
	}

	return nil
}

func (this *JsonParser) FormJsonToModel(ctx *context.Context, model interface{}) error {
	data := this.FormToJsonWithFieldsConfigs(ctx, nil)

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, &model)

}

func (this *JsonParser) FormJsonToModelWithCOnfigs(ctx *context.Context, model interface{}, configs []*FormJsonConfig) error {
	data := this.formToJsonWithFieldsConfigs(ctx, nil, configs)

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, &model)

}

func (this *JsonParser) FormToMap(ctx *context.Context, configs ...[]*FormJsonConfig) map[string]interface{} {
	if len(configs) > 0 {
		return this.FormToJsonWithConfigs(ctx, configs[0])
	} else {
		return this.FormToJson(ctx)
	}
}

func (this *JsonParser) FormToJson(ctx *context.Context) map[string]interface{} {
	return this.formToJsonWithFieldsConfigs(ctx, nil, nil)
}

func (this *JsonParser) FormToJsonWithFieldsConfigs(ctx *context.Context, configs map[string]string) map[string]interface{} {
	return this.formToJsonWithFieldsConfigs(ctx, configs, nil)
}

func (this *JsonParser) FormToJsonWithConfigs(ctx *context.Context, configs []*FormJsonConfig) map[string]interface{} {
	return this.formToJsonWithFieldsConfigs(ctx, nil, configs)
}

func (this *JsonParser) formToJsonWithFieldsConfigs(ctx *context.Context, configsMap map[string]string, configs []*FormJsonConfig) map[string]interface{} {
	data := ctx.Request.Form
	return this.UrlValuesToMapWithConfigs(data, configsMap, configs)
}

func (this *JsonParser) UrlValuesToMap(data url.Values) map[string]interface{} {
	return this.UrlValuesToMapWithConfigs(data, nil, nil)
}

func (this *JsonParser) UrlValuesToMapWithConfigs(data url.Values, configsMap map[string]string, configs []*FormJsonConfig) map[string]interface{} {

	jsonMap := make(map[string]interface{})

	findConfig := func(fieldName string) *FormJsonConfig {
		if configs != nil {
			for _, config := range configs {
				if config.FieldName == fieldName {
					return config
				}
			}
		}
		return nil
	}

	processValue := func(currentConfig *FormJsonConfig, value string) interface{} {
		if currentConfig != nil {

			if currentConfig.Parser != nil {
				return currentConfig.Parser(value)
			} else {
				switch currentConfig.FormatType {
				case FormatTypeFloat:

					if strings.Contains(value, ",") && strings.Contains(value, ".") {

						if strings.Index(value, ",") < strings.Index(value, ".") {
							return strings.Replace(value, ",", "", -1)
						}
						return strings.Replace(strings.Replace(value, ".", "", -1), ",", ".", -1)

					} else if strings.Contains(value, ",") {

						return strings.Replace(value, ",", ".", -1)

					} else if strings.TrimSpace(value) == "" {
						return "0.0"
					}

					return value

				case FormatTypeInt:

					if value == "" {
						return "0"
					}

					return value

				case FormatTypeDate:

					if len(value) > 0 {
						auxDate, _ := util.DateParse(currentConfig.Layout, value)
						jsonDateLayout := "2006-01-02T15:04:05-07:00"

						if !auxDate.IsZero() {
							return auxDate.Format(jsonDateLayout)
						}
					}
					return ""

				case FormatTypeBool:
					return value
				}
			}
		}

		return value
	}

	if configs == nil {
		configs = []*FormJsonConfig{}
	}

	for key, val := range configsMap {

		switch val {
		case "float":
			configs = append(configs, NewFormJsonConfig(key, FormatTypeFloat))
			break
		case "int":
			configs = append(configs, NewFormJsonConfig(key, FormatTypeInt))
			break
		case "bool":
			configs = append(configs, NewFormJsonConfig(key, FormatTypeBool))
			break
		default:

			if strings.Contains(val, "date:") {
				layout := strings.Replace(val, "date:", "", -1)

				configs = append(configs, NewFormJsonConfig(key, FormatTypeDate).SetLayout(layout))
			} else {
				logs.Error("Invalid format: %v. use [float | int | bool | date:DD/MM/YYYY]", val)
			}

			break

		}

	}

	for k, v := range data {

		if len(v) == 0 {
			continue
		}

		currentConfig := findConfig(k)

		if strings.Contains(k, ".") {
			keys := strings.Split(k, ".")

			parent := jsonMap

			for i, key := range keys {

				if currentConfig == nil {
					currentConfig = findConfig(key)
				}

				if _, ok := parent[key].(map[string]interface{}); !ok {
					parent[key] = make(map[string]interface{})
				}

				if i < len(keys)-1 {
					parent = parent[key].(map[string]interface{})
				} else {
					parent[key] = v[0]

					if parent[key] != nil {

						value := parent[key].(string)

						if currentConfig != nil {
							parent[key] = processValue(currentConfig, value)
						} else {
							parent[key] = value
						}
					}
				}
			}

		} else {

			//fmt.Println("k = ", k, " value = ", v[0])

			value := v[0]

			if currentConfig != nil {
				jsonMap[k] = processValue(currentConfig, value)
			} else {
				jsonMap[k] = value
			}

		}
	}

	return jsonMap
}

func (this *JsonParser) FormToModel(ctx *context.Context, model interface{}) error {
	return this.FormToModelWithFieldsConfigs(ctx, model, nil)
}

func (this *JsonParser) FormToModelWithFieldsConfigs(ctx *context.Context, model interface{}, configs map[string]string) error {
	jsonMap := this.FormToJsonWithFieldsConfigs(ctx, configs)
	return this.MapToModel(jsonMap, model)
}

func (this *JsonParser) MapToModel(data map[string]interface{}, model interface{}) error {
	jsonData, err := json.Marshal(data)

	if err != nil {
		return errors.New(fmt.Sprintf("error on FormToModel.json.Marshal: %v", err.Error()))
	}

	err = json.Unmarshal(jsonData, model)

	if err != nil {
		return errors.New(fmt.Sprintf("error on FormToModel.json.Unmarshal: %v", err.Error()))
	}

	return nil
}

func (this *JsonParser) StructToModel(data interface{}, model interface{}) error {
	jsonData, err := json.Marshal(data)

	if err != nil {
		return errors.New(fmt.Sprintf("error on FormToModel.json.Marshal: %v", err.Error()))
	}

	err = json.Unmarshal(jsonData, model)

	if err != nil {
		return errors.New(fmt.Sprintf("error on FormToModel.json.Unmarshal: %v", err.Error()))
	}

	return nil
}

func (this *JsonParser) GetJsonObject(json map[string]interface{}, key string) map[string]interface{} {

	if this.HasJsonKey(json, key) {
		opt, _ := json[key]
		if opt != nil {
			return opt.(map[string]interface{})
		}
	}

	return nil
}

func (this *JsonParser) GetJsonArray(json map[string]interface{}, key string) []map[string]interface{} {

	if this.HasJsonKey(json, key) {
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

func (this *JsonParser) GetJsonSimpleArray(json map[string]interface{}, key string) []interface{} {

	if this.HasJsonKey(json, key) {
		opt, _ := json[key]

		if array, ok := opt.([]interface{}); ok {
			return array
		}

	}

	return nil
}

func (this *JsonParser) GetArrayFromJson(json map[string]interface{}, key string) []interface{} {

	if this.HasJsonKey(json, key) {
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

func (this *JsonParser) GetJsonInt(json map[string]interface{}, key string) int {
	var val int

	if this.HasJsonKey(json, key) {
		if _, ok := json[key].(int); ok {
			val = json[key].(int)
		} else if _, ok := json[key].(int64); ok {
			val = int(json[key].(int64))
		} else if _, ok := json[key].(float64); ok {
			val = int(json[key].(float64))
		} else if _, ok := json[key].(float32); ok {
			val = int(json[key].(float32))
		} else {
			val, _ = strconv.Atoi(this.GetJsonString(json, key))
		}
	}

	return val
}

func (this *JsonParser) GetJsonInt64(json map[string]interface{}, key string) int64 {

	var val int

	if this.HasJsonKey(json, key) {
		if _, ok := json[key].(int); ok {
			val = json[key].(int)
		} else if _, ok := json[key].(int64); ok {
			val = int(json[key].(int64))
		} else if _, ok := json[key].(float64); ok {
			val = int(json[key].(float64))
		} else if _, ok := json[key].(float32); ok {
			val = int(json[key].(float32))
		} else {
			val, _ = strconv.Atoi(this.GetJsonString(json, key))
		}
	}

	return int64(val)
}

func (this *JsonParser) GetJsonFloat32(json map[string]interface{}, key string) float32 {

	var val float32

	if this.HasJsonKey(json, key) {
		if _, ok := json[key].(float32); ok {
			val = json[key].(float32)
		} else if _, ok := json[key].(float64); ok {
			val = float32(json[key].(float64))
		} else if _, ok := json[key].(int64); ok {
			val = float32(json[key].(int64))
		} else if _, ok := json[key].(int); ok {
			val = float32(json[key].(int))
		} else {
			v, _ := strconv.ParseFloat(this.GetJsonString(json, key), 32)
			val = float32(v)
		}
	}

	return float32(val)
}

func (this *JsonParser) GetJsonFloat64(json map[string]interface{}, key string) float64 {

	var val float64

	if this.HasJsonKey(json, key) {
		if _, ok := json[key].(float64); ok {
			val = json[key].(float64)
		} else if _, ok := json[key].(float32); ok {
			val = float64(json[key].(float32))
		} else if _, ok := json[key].(int64); ok {
			val = float64(json[key].(int64))
		} else if _, ok := json[key].(int); ok {
			val = float64(json[key].(int))
		} else {
			v, _ := strconv.ParseFloat(this.GetJsonString(json, key), 64)
			val = float64(v)
		}
	}

	return float64(val)
}

func (this *JsonParser) GetJsonBool(json map[string]interface{}, key string) bool {

	var val bool

	if this.HasJsonKey(json, key) {
		if _, ok := json[key].(bool); ok {
			val = json[key].(bool)
		} else {
			val, _ = strconv.ParseBool(this.GetJsonString(json, key))
		}
	}

	return val
}

func (this *JsonParser) GetJsonString(json map[string]interface{}, key string) string {

	var val string

	if !this.HasJsonKey(json, key) {
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

func (this *JsonParser) JsonInterfaceToInt64(item interface{}) int64 {

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

func (this *JsonParser) GetJsonDate(json map[string]interface{}, key string, layout string) time.Time {
	date, _ := time.ParseInLocation(layout, this.GetJsonString(json, key), this.DefaultLocation)
	return date
}

func (this *JsonParser) HasJsonKey(json map[string]interface{}, key string) bool {
	if _, ok := json[key]; ok {
		return true
	}
	return false
}
