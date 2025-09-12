package json

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-io/result"
	goioutil "github.com/mobilemindtech/go-io/util"
	ioutil "github.com/mobilemindtech/go-io/util"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/v2/lists"
	"github.com/mobilemindtech/go-utils/v2/optional"
	"github.com/mobilemindtech/go-utils/v2/try"
)

type JsonWriter interface {
	Write(data map[string]interface{})
}

type JsonReader interface {
	Reader(data map[string]interface{})
}

type Converter func(j *Json)

type ParserConfig struct {
	Converter         Converter
	UseDefaultEncoder bool
	UseCamelCase      bool
	Debug             bool
	TagNames          []string
}

// NewParserConfig parser config with default config using tagNames = [jsonp] and NewJSON encode/decoder
func NewParserConfig() *ParserConfig {
	return &ParserConfig{TagNames: []string{}}
}

// NewParserConfigWithDefaultJsonTagName parser config with default config using tagNames = [json, jsonp] and NewJSON encode/decoder
func NewParserConfigWithDefaultJsonTagName() *ParserConfig {
	return &ParserConfig{TagNames: []string{"json"}}
}

func (this *ParserConfig) GetTagNames() []string {
	if len(this.TagNames) == 0 {
		this.TagNames = []string{}
	}
	return this.TagNames
}

func (this *ParserConfig) AddTagNames(tag string) *ParserConfig {
	if len(this.TagNames) == 0 {
		this.TagNames = []string{}
	}
	this.TagNames = append(this.TagNames, tag)
	return this
}

type Parser[T any] struct {
	cfg *ParserConfig
}

// NewParser new parse with empty configs. Use NewJSON api to encode/decode
func NewParser[T any](cfgs ...*ParserConfig) *Parser[T] {

	cfg := &ParserConfig{}
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}

	return &Parser[T]{cfg}
}

// NewParserDefault new parse with empty configs. Use golang json.Marshal to encode/decode
func NewParserDefault[T any]() *Parser[T] {
	return NewParser[T](&ParserConfig{UseDefaultEncoder: true})
}

func (this *Parser[T]) ConfigureTagName(name string) *Parser[T] {
	this.cfg.AddTagNames(name)
	return this
}

// UseDefaultEncoder Set to use golang json.Marshal to encode/decode
func (this *Parser[T]) UseDefaultEncoder() *Parser[T] {
	return this.SetUseDefaultEncoder(true)
}

func (this *Parser[T]) Debug() *Parser[T] {
	this.cfg.Debug = true
	return this
}

// UseCamelCase set to usse calmelcase json field names to NewJSON encoder. Default is snackcase.
func (this *Parser[T]) UseCamelCase() *Parser[T] {
	this.cfg.UseCamelCase = true
	return this
}

// UseDefaultEncoder Set to use golang json.Marshal to encode/decode
func (this *Parser[T]) SetUseDefaultEncoder(b bool) *Parser[T] {
	this.cfg.UseDefaultEncoder = true
	return this
}

func (this *Parser[T]) AddConverter(c Converter) *Parser[T] {
	this.cfg.Converter = c
	return this
}

func (this *Parser[T]) Parse(raw []byte) *optional.Optional[T] {
	entity := goioutil.NewOf[T]()
	return this.ParseInto(raw, entity)
}

func (this *Parser[T]) ParseFormTo(form url.Values, entity T) *optional.Optional[T] {
	return this.ParseJsonInto(NewFromUrlValues(form), entity)
}

func (this *Parser[T]) ParseForm(form url.Values) *optional.Optional[T] {
	entity := goioutil.NewOf[T]()
	return this.ParseJsonInto(NewFromUrlValues(form), entity)
}

func (this *Parser[T]) ParseJson(j *Json) *optional.Optional[T] {
	entity := goioutil.NewOf[T]()
	return this.ParseJsonInto(j, entity)
}

func (this *Parser[T]) ParseInto(raw []byte, entity T) *optional.Optional[T] {
	j, err := NewFromBytes(raw)

	if err != nil {
		return optional.OfFail[T](err)
	}

	return this.ParseJsonInto(j, entity)
}

func (this *Parser[T]) ParseJsonInto(j *Json, entity T) *optional.Optional[T] {

	if this.cfg.Converter != nil {
		this.cfg.Converter(j)
	}

	var err error

	if this.cfg.Debug {
		logs.Debug("JSON DATA = %v", j.data)
	}

	if this.cfg.UseDefaultEncoder {

		newJsonData, err := json.Marshal(j.data)

		if err != nil {
			return optional.OfFail[T](err)
		}

		if this.cfg.Debug {
			logs.Debug("JSON RAW = %v", string(newJsonData))
		}

		err = json.Unmarshal(newJsonData, entity)

	} else {

		jsn := NewJSON()
		jsn.CamelCase = this.cfg.UseCamelCase
		for _, tagName := range this.cfg.GetTagNames() {
			jsn.ConfigureTagName(tagName)
		}
		err = jsn.DecodeFromMap(j.data, entity)

	}

	if err != nil {
		return optional.OfFail[T](err)
	}

	return optional.OfSome[T](entity)
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

		switch opt.(type) {
		case []interface{}:
			var arr []int
			for _, it := range opt.([]interface{}) {
				switch it.(type) {
				case int:
					arr = append(arr, it.(int))
					break
				case int64:
					arr = append(arr, int(it.(int64)))
					break
				case float32:
					arr = append(arr, int(it.(float32)))
					break
				case float64:
					arr = append(arr, int(it.(float64)))
					break
				case string:
					arr = append(arr, support.StrToInt(it.(string)))
					break
				default:
					panic(fmt.Errorf("can't parse %v item of type %v of int", key, reflect.TypeOf(it)))
				}
			}
			return arr
		case []int:
			return opt.([]int)
		case []int64:
			return lists.Map(opt.([]int64), func(i int64) int { return int(i) })
			//default:
			//	panic(fmt.Errorf("can't parse %v of type %v to array of int", key, reflect.TypeOf(opt)))
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

		switch opt.(type) {
		case []string:
			return opt.([]string)
		case []interface{}:
			var arr []string
			for _, it := range opt.([]interface{}) {
				switch it.(type) {
				case string:
					arr = append(arr, it.(string))
					break
				default:
					arr = append(arr, fmt.Sprintf("%v", it))
				}
			}
			return arr
			//default:
			//	panic(fmt.Errorf("can't parse %v of type %v to array of string", key, reflect.TypeOf(opt)))
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

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func MarshalResult(v interface{}) *result.Result[[]byte] {
	return result.Try[[]byte](func() ([]byte, error) {
		return json.Marshal(v)
	})
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func Unmarshal(data []byte, m interface{}) error {
	return json.Unmarshal(data, m)
}

func UnmarshalResult[T any](data []byte) *result.Result[T] {
	val := ioutil.NewOf[T]()
	err := json.Unmarshal(data, val)
	return result.OfErrorOrValue(err, val)
}
