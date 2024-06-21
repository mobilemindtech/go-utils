package json

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-io/result"
	"github.com/mobilemindtec/go-utils/app/util"
	"github.com/mobilemindtec/go-utils/support"
	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/mobilemindtec/go-utils/v2/try"
)

type JsonWriter interface {
	Write(data map[string]interface{})
}

type JsonReader interface {
	Reader(data map[string]interface{})
}

type Converter func(j *Json)

type Parser[T any] struct {
	converter         Converter
	useDefaultEncoder bool
	useCamelCase      bool
	debug             bool
}

func NewParser[T any]() *Parser[T] {
	return &Parser[T]{}
}

func NewParserDefault[T any]() *Parser[T] {
	return &Parser[T]{useDefaultEncoder: true}
}

func (this *Parser[T]) UseDefaultEncoder() *Parser[T] {
	return this.SetUseDefaultEncoder(true)
}

func (this *Parser[T]) Debug() *Parser[T] {
	this.debug = true
	return this
}

func (this *Parser[T]) UseCamelCase() *Parser[T] {
	this.useCamelCase = true
	return this
}

func (this *Parser[T]) SetUseDefaultEncoder(b bool) *Parser[T] {
	this.useDefaultEncoder = b
	return this
}

func (this *Parser[T]) AddConverter(c Converter) *Parser[T] {
	this.converter = c
	return this
}

func (this *Parser[T]) Parse(raw []byte) *optional.Optional[*T] {
	var entity T
	return this.ParseInto(raw, &entity)
}

func (this *Parser[T]) ParseFormTo(form url.Values, entity *T) *optional.Optional[*T] {
	return this.ParseJsonInto(NewFromUrlValues(form), entity)
}

func (this *Parser[T]) ParseForm(form url.Values) *optional.Optional[*T] {
	var entity T
	return this.ParseJsonInto(NewFromUrlValues(form), &entity)
}

func (this *Parser[T]) ParseJson(j *Json) *optional.Optional[*T] {
	var entity T
	return this.ParseJsonInto(j, &entity)
}

func (this *Parser[T]) ParseInto(raw []byte, entity *T) *optional.Optional[*T] {
	j, err := NewFromBytes(raw)

	if err != nil {
		return optional.OfFail[*T](err)
	}

	return this.ParseJsonInto(j, entity)
}

func (this *Parser[T]) ParseJsonInto(j *Json, entity *T) *optional.Optional[*T] {

	if this.converter != nil {
		this.converter(j)
	}

	var err error

	if this.debug {
		logs.Debug("JSON DATA = %v", j.data)
	}
	
	if this.useDefaultEncoder {


		newJsonData, err := json.Marshal(j.data)


		if err != nil {
			return optional.OfFail[*T](err)
		}

		if this.debug {
			logs.Debug("JSON RAW = %v", string(newJsonData))
		}

		err = json.Unmarshal(newJsonData, entity)

	} else {

		jsn := NewJSON()
		jsn.CamelCase = this.useCamelCase
		err = jsn.DecodeFromMap(j.data, entity)


	}

	if err != nil {
		return optional.OfFail[*T](err)
	}

	return optional.OfSome[*T](entity)
}

type Json struct {
	support.JsonParser
	data map[string]interface{}
	raw  []byte
}

func NewFromMap(d map[string]interface{}) *Json {
	j := new(Json)
	j.data = d
	return j
}

func NewFromUrlValues(form url.Values) *Json {
	data := support.NewJsonParser().UrlValuesToMap(form)
	return NewFromMap(data)
}

func NewFromBytes(raw []byte) (*Json, error) {
	j := new(Json)
	data, err := j.JsonBytesToMap(raw)

	if err != nil {
		return nil, err
	}

	j.data = data
	j.raw = raw
	return j, nil
}

func New(raw []byte) interface{} {
	j, err := NewFromBytes(raw)

	if err != nil {
		return optional.NewFail(err)
	}
	return optional.NewSome(j)
}

func Of(raw []byte) *optional.Optional[*Json] {
	return try.Of(func() (*Json, error) {
		return NewFromBytes(raw)
	})
}

func Try(raw []byte) *result.Result[*Json] {
	return result.Try(func() (*Json, error) {
		return NewFromBytes(raw)
	})
}

func NewEmpty() *Json {
	return &Json{data: make(map[string]interface{})}
}

func (this *Json) GetData() map[string]interface{} {
	return this.data
}

func (this *Json) Set(key string, value interface{}) {
	this.data[key] = value
}

func (this *Json) SetNested(key string, nestedKey string, value interface{}) {
	this.GetObject(key).Set(nestedKey, value)
}

func (this *Json) OptObject(key string, def *Json) *Json {
	r := this.GetObject(key)
	if r == nil {
		r = def
	}
	return r
}

func (this *Json) GetObject(key string) *Json {

	if this.HasKey(key) {
		opt, _ := this.data[key]
		if opt != nil {
			return NewFromMap(opt.(map[string]interface{}))
		}

	}

	return nil
}

func (this *Json) OptObjectArray(key string, def []*Json) []*Json {
	r := this.GetObjectArray(key)
	if r == nil {
		r = def
	}
	return r
}

func (this *Json) GetObjectArray(key string) []*Json {

	if this.HasKey(key) {
		opt, _ := this.data[key]

		j := []*Json{}

		if array, ok := opt.([]interface{}); ok {
			for _, it := range array {
				if p, ok := it.(map[string]interface{}); ok {
					j = append(j, NewFromMap(p))
				}
			}
		}

		return j
	}

	return nil
}

func (this *Json) OptArray(key string, def []interface{}) []interface{} {
	r := this.GetArray(key)
	if r == nil {
		r = def
	}
	return r
}

func (this *Json) GetArray(key string) []interface{} {

	if this.HasKey(key) {
		opt, _ := this.data[key]

		if array, ok := opt.([]interface{}); ok {
			return array
		}

	}

	return nil
}

func (this *Json) GetArrayOrEmpty(key string) []interface{} {
	empty := []interface{}{}
	return this.OptArray(key, empty)
}

func (this *Json) GetArrayOfInt(key string) []int {

	if this.HasKey(key) {
		opt, _ := this.data[key]

		if array, ok := opt.([]int); ok {
			return array
		}

	}

	return nil
}

func (this *Json) OptArrayOfInt(key string, def []int) []int {
	r := this.GetArrayOfInt(key)
	if r == nil {
		r = def
	}
	return r
}

func (this *Json) GetArrayOfString(key string) []string {

	if this.HasKey(key) {
		opt, _ := this.data[key]

		logs.Debug("opt = %v, type = %v", opt, reflect.TypeOf(opt))

		if array, ok := opt.([]string); ok {
			return array
		}

	}

	return nil
}

func (this *Json) GetArrayOfStringOrEmpty(key string) []string {
	empty := []string{}
	return this.OptArrayOfString(key, empty)
}

func (this *Json) OptArrayOfString(key string, def []string) []string {
	r := this.GetArrayOfString(key)
	if r != nil {
		r = def
	}
	return r
}

func (this *Json) GetArrayOfIntOrEmpty(key string) []int {
	empty := []int{}
	return this.OptArrayOfInt(key, empty)
}

func (this *Json) OptInt(key string, def int) int {
	if !this.HasKey(key) {
		return def
	}
	return this.GetInt(key)

}

func (this *Json) GetInt(key string) int {
	var val int

	if this.HasKey(key) {
		if v, ok := this.data[key].(int); ok {
			val = v
		} else if v, ok := this.data[key].(int64); ok {
			val = int(v)
		} else if v, ok := this.data[key].(float64); ok {
			val = int(v)
		} else if v, ok := this.data[key].(float32); ok {
			val = int(v)
		} else {
			val, _ = strconv.Atoi(this.GetString(key))
		}
	}

	return val
}

func (this *Json) OptInt64(key string, def int64) int64 {
	if !this.HasKey(key) {
		return def
	}
	return this.GetInt64(key)
}

func (this *Json) GetInt64(key string) int64 {

	var val int

	if this.HasKey(key) {
		if v, ok := this.data[key].(int); ok {
			val = v
		} else if v, ok := this.data[key].(int64); ok {
			val = int(v)
		} else if v, ok := this.data[key].(float64); ok {
			val = int(v)
		} else if v, ok := this.data[key].(float32); ok {
			val = int(v)
		} else {
			val, _ = strconv.Atoi(this.GetString(key))
		}
	}

	return int64(val)
}

func (this *Json) OptFloat(key string, def float32) float32 {
	if !this.HasKey(key) {
		return def
	}
	return this.GetFloat(key)
}

func (this *Json) GetFloat(key string) float32 {

	var val float32

	if this.HasKey(key) {
		if v, ok := this.data[key].(float32); ok {
			val = v
		} else if v, ok := this.data[key].(float64); ok {
			val = float32(v)
		} else if v, ok := this.data[key].(int64); ok {
			val = float32(v)
		} else if v, ok := this.data[key].(int); ok {
			val = float32(v)
		} else {
			v, _ := strconv.ParseFloat(this.GetString(key), 32)
			val = float32(v)
		}
	}

	return val
}

func (this *Json) OptFloat64(key string, def float64) float64 {
	if !this.HasKey(key) {
		return def
	}
	return this.GetFloat64(key)
}

func (this *Json) GetFloat64(key string) float64 {

	var val float64

	if this.HasKey(key) {
		if v, ok := this.data[key].(float64); ok {
			val = v
		} else if v, ok := this.data[key].(float32); ok {
			val = float64(v)
		} else if v, ok := this.data[key].(int64); ok {
			val = float64(v)
		} else if v, ok := this.data[key].(int); ok {
			val = float64(v)
		} else {
			v, _ := strconv.ParseFloat(this.GetString(key), 64)
			val = float64(v)
		}
	}

	return val
}

func (this *Json) OptBool(key string, def bool) bool {
	if !this.HasKey(key) {
		return def
	}
	return this.GetBool(key)
}

func (this *Json) GetBool(key string) bool {

	var val bool

	if this.HasKey(key) {
		if v, ok := this.data[key].(bool); ok {
			val = v
		} else {
			val, _ = strconv.ParseBool(this.GetString(key))
		}
	}

	return val
}

func (this *Json) OptString(key string, def string) string {
	if !this.HasKey(key) {
		return def
	}
	return this.GetString(key)
}

func (this *Json) GetString(key string) string {

	var val string

	if this.HasKey(key) {
		if v, ok := this.data[key].(string); ok {
			val = v
			if val == "null" || val == "undefined" {
				return val
			}
		}
	}

	return val
}

func (this *Json) OptTime(key string, layout string, def time.Time) time.Time {
	if !this.HasKey(key) {
		return def
	}
	return this.GetTime(key, layout)
}

func (this *Json) OptTimeWithLocation(key string, layout string, loc *time.Location, def time.Time) time.Time {
	if !this.HasKey(key) {
		return def
	}
	return this.GetTimeWithLocation(key, layout, loc)
}

func (this *Json) GetTime(key string, layout string) time.Time {
	return this.GetTimeWithLocation(key, layout, util.GetDefaultLocation())
}

func (this *Json) GetTimeWithLocation(key string, layout string, loc *time.Location) time.Time {
	date, _ := time.ParseInLocation(layout, this.GetString(key), loc)
	return date
}

func (this *Json) HasNotKeys(keys ...string) bool {
	return !this.HasKeys(keys...)
}
func (this *Json) HasKeys(keys ...string) bool {
	for _, key := range keys {
		if !this.HasKey(key) {
			return false
		}
	}
	return true
}

func (this *Json) HasNotKey(key string) bool {
	return !this.HasKey(key)
}

func (this *Json) HasKey(key string) bool {
	if _, ok := this.data[key]; ok {
		return true
	}
	return false
}

func (this *Json) LogRaw() {
	logs.Info("JSON RAW: %v", string(this.raw))
}

func (this *Json) LogData() {
	logs.Info("JSON DATA: %v", this.data)
}

func (this *Json) LogAll() {
	this.LogRaw()
	this.LogData()
}
