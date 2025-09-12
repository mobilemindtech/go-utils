package db

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/mobilemindtech/go-utils/app/util"
	"strconv"
	"time"
)

type DbResult = map[string]interface{}
type DbResultSet = []DbResult

type RawQuery struct {
	Query string
	Args  []interface{}

	Session *Session

	Error        error
	RowsAffected int64

	values     []orm.Params
	valuesList []orm.ParamsList
	valuesFlat orm.ParamsList
}

func NewRawQuerySession(session *Session) *RawQuery {
	return &RawQuery{Session: session}
}

func NewRawQuery(session *Session, query string) *RawQuery {
	q := &RawQuery{Session: session}
	return q.WithQuery(query)
}

func NewRawQueryArgs(session *Session, query string, args ...interface{}) *RawQuery {
	q := &RawQuery{Session: session}
	return q.WithQuery(query).WithArgs(args...)
}

func (this *RawQuery) HasError() bool {
	return this.Error != nil
}

func (this *RawQuery) WithSession(session *Session) *RawQuery {
	this.Session = session
	return this
}

func (this *RawQuery) WithQuery(query string) *RawQuery {
	this.Query = query
	return this
}

func (this *RawQuery) WithArgs(args ...interface{}) *RawQuery {
	this.Args = args
	return this
}

func (this *RawQuery) RawSeter() orm.RawSeter {
	return this.Session.GetDb().Raw(this.Query, this.Args...)
}

func (this *RawQuery) Execute() (int64, error) {
	res, err := this.Session.GetDb().Raw(this.Query, this.Args...).Exec()

	if err != nil {
		this.Error = err
		return 0, err
	}

	this.RowsAffected, this.Error = res.RowsAffected()

	return this.RowsAffected, this.Error
}

func (this *RawQuery) ToResutl(result interface{}) error {
	this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).QueryRow(result)
	return this.Error
}

func (this *RawQuery) ToResutls(results interface{}) error {
	this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).QueryRows(results)
	return this.Error
}

func (this *RawQuery) Values() *RawQuery {
	this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).Values(&this.values)
	return this
}

func (this *RawQuery) ExecAsValuesMap() ([]map[string]interface{}, error) {
	this.Values()
	if this.HasError() {
		return nil, this.Error
	}
	return this.GetValues(), nil
}

func (this *RawQuery) ValuesList() *RawQuery {
	this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).ValuesList(&this.valuesList)
	return this
}

func (this *RawQuery) ExecAsValuesList() ([][]interface{}, error) {
	this.ValuesList()
	if this.HasError() {
		return nil, this.Error
	}
	return this.GetValuesList(), nil
}

func (this *RawQuery) ValuesFlat() *RawQuery {
	this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).ValuesFlat(&this.valuesFlat)
	return this
}
func (this *RawQuery) ExecAsValuesFlat() ([]interface{}, error) {
	this.ValuesFlat()
	if this.HasError() {
		return nil, this.Error
	}
	return this.GetValuesFlat(), nil
}

func (this *RawQuery) GetValuesFlat() []interface{} {
	results := []interface{}{}

	for _, x := range this.valuesFlat {
		results = append(results, x)
	}

	return results
}

func (this *RawQuery) GetValues() []map[string]interface{} {
	results := []map[string]interface{}{}

	for _, item := range this.values {

		itemMap := make(map[string]interface{})

		for k, v := range item {
			itemMap[k] = v
		}

		results = append(results, itemMap)
	}

	return results

}

func (this *RawQuery) GetValuesList() [][]interface{} {

	results := [][]interface{}{}

	for _, item := range this.valuesList {

		newItem := []interface{}{}

		for _, v := range item {
			newItem = append(newItem, v)
		}

		results = append(results, newItem)
	}

	return results
}

func (this *RawQuery) ExecAsSlice() ([]interface{}, error) {
	this.ValuesFlat()
	if this.HasError() {
		return nil, this.Error
	}
	return this.GetValuesFlat(), nil
}

func (this *RawQuery) Exec() (DbResultSet, error) {
	this.Values()
	if this.HasError() {
		return nil, this.Error
	}
	return this.GetValues(), nil
}

func (this *RawQuery) ExecAsSliceOSlice() ([][]interface{}, error) {
	this.ValuesList()
	if this.HasError() {
		return nil, this.Error
	}
	return this.GetValuesList(), nil
}

func ValueToInt(val interface{}) int {
	switch val.(type) {
	case int:
		return val.(int)
	case int64:
		return int(val.(int64))
	default:
		v, _ := strconv.Atoi(val.(string))
		return v
	}
}

func ValueToInt64(val interface{}) int64 {
	switch val.(type) {
	case int:
		return int64(val.(int))
	case int64:
		return val.(int64)
	default:
		v, _ := strconv.Atoi(val.(string))
		return int64(v)
	}
}

func ValueToTime(val interface{}) (time.Time, error) {
	return ValueToTimeWithPattern(val, util.DateTimeDbLayout)
}

func ValueToTimeWithPattern(val interface{}, pattern string) (time.Time, error) {
	switch val.(type) {
	case time.Time:
		return val.(time.Time), nil
	case string:
		return util.DateParse(pattern, val.(string))
	default:
		return time.Time{}, fmt.Errorf("can't parse %v to time", val)
	}
}

type RawQueryTransformer[T any] struct {
	RawQuery
	ValuesTransFn func(map[string]interface{}) (T, error)
	ListTransFn   func([]interface{}) (T, error)
}

func NewRawQueryTransformerSession[T any](session *Session) *RawQueryTransformer[T] {
	q := &RawQueryTransformer[T]{}
	q.WithSession(session)
	return q
}

func NewRawQueryTransformer[T any](session *Session, query string) *RawQueryTransformer[T] {
	q := &RawQueryTransformer[T]{}
	q.WithSession(session).WithQuery(query)
	return q
}

func NewRawQueryTransformerArgs[T any](session *Session, query string, args ...interface{}) *RawQueryTransformer[T] {
	q := &RawQueryTransformer[T]{}
	q.WithSession(session).WithQuery(query).WithArgs(args...)
	return q
}

func (this *RawQueryTransformer[T]) WithValuesTransformer(tr func(map[string]interface{}) (T, error)) *RawQueryTransformer[T] {
	this.ValuesTransFn = tr
	return this
}

func (this *RawQueryTransformer[T]) WithListTransformer(tr func([]interface{}) (T, error)) *RawQueryTransformer[T] {
	this.ListTransFn = tr
	return this
}

func (this *RawQueryTransformer[T]) List() ([]T, error) {

	if this.ListTransFn == nil && this.ValuesTransFn == nil {
		return nil, fmt.Errorf("use values or value list transformer function")
	}

	if this.ListTransFn != nil && this.ValuesTransFn != nil {
		return nil, fmt.Errorf("use values or value list transformer function")
	}

	results := []T{}

	if this.ValuesTransFn != nil {

		values := this.GetValues()

		for _, it := range values {
			result, err := this.ValuesTransFn(it)

			if err != nil {
				return nil, err
			}

			results = append(results, result)
		}

	} else {

		values := this.GetValuesList()

		for _, item := range values {
			result, err := this.ListTransFn(item)

			if err != nil {
				return nil, err
			}

			results = append(results, result)
		}

	}

	return results, nil
}

func (this *RawQueryTransformer[T]) First() (T, error) {

	results, err := this.List()
	var r T

	if err != nil {
		return r, err
	}

	if len(results) > 0 {
		return results[0], nil
	}

	return r, nil

}
