package features

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/support"
)

type WebParser struct {
	support.JsonParser
	base trait.WebBaseInterface
}

func (this *WebParser) InitWebParser(base trait.WebBaseInterface) {
	this.base = base
	this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")
}

func (this *WebParser) OnParseForm(entity interface{}) {
	if err := this.base.GetBeegoController().ParseForm(entity); err != nil {
		logs.Error("*******************************************")
		logs.Error("***** ERROR parse FORM to JSON: %v", err.Error())
		logs.Error("*******************************************")
		this.base.GetBeegoController().Abort("500")
	}
}

func (this *WebParser) OnJsonParseForm(entity interface{}) {
	this.Form2Json(entity)
}

func (this *WebParser) OnJsonParseFormWithFieldsConfigs(entity interface{}, configs map[string]string) {
	this.Form2JsonWithCnf(entity, configs)
}

func (this *WebParser) Form2Json(entity interface{}) {
	this.Form2JsonWithCnf(entity, nil)
}

func (this *WebParser) Form2JsonWithCnf(entity interface{}, configs map[string]string) {
	if err := this.FormToModelWithFieldsConfigs(this.base.GetBeegoController().Ctx, entity, configs); err != nil {
		logs.Error("*******************************************")
		logs.Error("***** ERROR parse FORM to JSON: %v ", err.Error())
		logs.Error("*******************************************")
		this.base.GetBeegoController().Abort("500")
	}
}

func (this *WebParser) ParamParseMoney(s string) float64 {
	return this.ParamParseFloat(s)
}

// remove ,(virgula) do valor em params que vem como val de input com jquery money
// exemplo 45,000.00 vira 45000.00
func (this *WebParser) ParamParseFloat(s string) float64 {
	var semic string = ","
	replaced := strings.Replace(s, semic, "", -1) // troca , por espaço
	precoFloat, err := strconv.ParseFloat(replaced, 64)
	var returnValue float64
	if err == nil {
		returnValue = precoFloat
	} else {
		logs.Error("*******************************************")
		logs.Error("****** ERROR parse string to float64 for stringv", s)
		logs.Error("*******************************************")
		this.base.GetBeegoController().Abort("500")
	}

	return returnValue
}

func (this *WebParser) OnParseJson(entity interface{}) {
	if err := this.JsonToModel(this.base.GetBeegoController().Ctx, entity); err != nil {
		logs.Error("*******************************************")
		logs.Error("***** ERROR on parse json ", err.Error())
		logs.Error("*******************************************")
		this.base.GetBeegoController().Abort("500")
	}
}

func (this *WebParser) StringToInt(text string) int {
	val, _ := strconv.Atoi(text)
	return val
}

func (this *WebParser) StringToInt64(text string) int64 {
	val, _ := strconv.ParseInt(text, 10, 64)
	return val
}

func (this *WebParser) IntToString(val int) string {
	return fmt.Sprintf("%v", val)
}

func (this *WebParser) Int64ToString(val int64) string {
	return fmt.Sprintf("%v", val)
}

func (this *WebParser) GetId() int64 {
	return this.GetIntParam(":id")
}

func (this *WebParser) GetParam(key string) string {

	if !strings.HasPrefix(key, ":") {
		key = fmt.Sprintf(":%v", key)
	}

	return this.base.GetBeegoController().Ctx.Input.Param(key)
}

func (this *WebParser) GetStringParam(key string) string {
	return this.GetParam(key)
}

func (this *WebParser) GetIntParam(key string) int64 {
	id := this.GetParam(key)
	intid, _ := strconv.ParseInt(id, 10, 64)
	return intid
}

func (this *WebParser) GetInt32Param(key string) int {
	val := this.GetParam(key)
	intid, _ := strconv.Atoi(val)
	return intid
}

func (this *WebParser) GetBoolParam(key string) bool {
	val := this.GetParam(key)
	return val == "true"
}

func (this *WebParser) GetIntByKey(key string) int64 {
	val := this.base.GetBeegoController().Ctx.Input.Query(key)
	intid, _ := strconv.ParseInt(val, 10, 64)
	return intid
}

func (this *WebParser) GetBoolByKey(key string) bool {
	val := this.base.GetBeegoController().Ctx.Input.Query(key)
	boolean, _ := strconv.ParseBool(val)
	return boolean
}

func (this *WebParser) GetCheckbox(key string) bool {
	val := this.base.GetBeegoController().GetString(key)
	return strings.ToLower(val) == "on"
}

func (this *WebParser) GetStringByKey(key string) string {
	return this.base.GetBeegoController().Ctx.Input.Query(key)
}

func (this *WebParser) GetDateByKey(key string) (time.Time, error) {
	date := this.base.GetBeegoController().Ctx.Input.Query(key)
	return this.ParseDate(date)
}

func (this *WebParser) ParseDateByKey(key string, layout string) (time.Time, error) {
	date := this.base.GetBeegoController().Ctx.Input.Query(key)
	return time.ParseInLocation(layout, date, this.DefaultLocation)
}

// deprecated
func (this *WebParser) ParseDate(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateBrLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebParser) ParseDateTime(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateTimeBrLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebParser) ParseJsonDate(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateTimeDbLayout, date, this.DefaultLocation)
}

func (this *WebParser) NormalizePageSortKey(key string) string {
	if strings.Contains(key, ".") {
		return strings.Replace(key, ".", "__", -1)
	}
	return key
}

func (this *WebParser) CheckboxToBool(key string) bool {
	arr := this.base.GetBeegoController().Ctx.Request.Form[key]
	return len(arr) > 0
}
