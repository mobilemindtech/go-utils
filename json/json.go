package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/app/util"
	"github.com/mobilemindtec/go-utils/support"
)

const (
	TimestampLayout  string = "2006-01-02T15:04:05-07:00"
	TimestampLayout2        = "2006-01-02T15:04:05.000'Z'"
	DateLayout              = "2006-02-01"
	DateTimeLayout          = "2006-02-01 15:04:05"
	TimeLayout              = "10:25:05"
	timeStringKind          = "time.Time"
	tagName                 = "jsonp"
)

type JSON struct {
	support.JsonParser
	Debug           bool
	DateFormat      string
	DateTimeFormat  string
	TimeFormat      string
	TimestampFormat string

	DebugParse  bool
	DebugFormat bool
	DateLayouts []string
}

func NewJSON() *JSON {
	return &JSON{
		DateFormat:      DateLayout,
		DateTimeFormat:  DateTimeLayout,
		TimeFormat:      TimeLayout,
		TimestampFormat: TimestampLayout,
		DateLayouts:     []string{TimestampLayout, TimestampLayout2, DateTimeLayout, DateTimeLayout, TimeLayout},
	}
}

func (this *JSON) EncodeToString(obj interface{}) (string, error) {
	result, err := this.Encode(obj)

	if err != nil {
		return "", err
	}

	return string(result), err
}

func (this *JSON) Encode(obj interface{}) ([]byte, error) {

	data, err := this.ParseObj(obj)

	if err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(data, "", " ")

	if this.DebugFormat {
		logs.Debug("JSON = ", string(result))
	}

	return result, err
}

func (this *JSON) ParseObj(obj interface{}) (interface{}, error) {
	refValue := reflect.ValueOf(obj)
	//fullValue := refValue
	//fullType := fullValue.Type()

	//logs.Debug("1 fullType ", fullType, " fullValue ", fullValue)
	kind := reflect.TypeOf(obj).Kind()

	switch kind {

	case reflect.Map:

		//logs.Info("is map")

		jsonResult := make(map[string]interface{})

		for _, key := range refValue.MapKeys() {
			v := refValue.MapIndex(key)
			r, err := this.ParseObj(v.Interface())
			if err != nil {
				return nil, err
			}
			jsonResult[key.Interface().(string)] = r
		}

		return jsonResult, nil

	case reflect.Slice:

		//logs.Info("is slice")
		jsonResult := []interface{}{}
		lst := reflect.Indirect(refValue)
		for i := 0; i < lst.Len(); i++ {
			r, err := this.ParseObj(lst.Index(i))
			if err != nil {
				return nil, err
			}
			jsonResult = append(jsonResult, r)
		}
		return jsonResult, nil

	case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:
		//logs.Info("is simple type")
		return obj, nil
	default:
		//logs.Info("is unkonw type")
		return this.ToMap(obj)
	}
}

func (this *JSON) ToMap(obj interface{}) (map[string]interface{}, error) {
	// value e type of pointer

	defer func() {
		if r := recover(); r != nil {
			logs.Debug("JSON TO MAP ERROR: ", r, ", OBJ = ", obj)
			panic(r)
		}
	}()

	refValue := reflect.ValueOf(obj)
	fullValue := refValue
	fullType := fullValue.Type()

	//logs.Debug("1 fullType ", fullType, " fullValue ", fullValue)

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		//logs.Debug("IS PTR")
		fullValue = refValue.Elem()
		fullType = refValue.Elem().Type()
	}

	//logs.Debug("2 fullType ", fullType, " fullValue ", fullValue)

	if fullValue.Kind() == reflect.Interface {
		//logs.Debug("IS INTERFACE")
		fullValue = refValue.Elem().Elem()
		fullType = refValue.Elem().Elem().Type()

		if fullValue.Kind() == reflect.Ptr {
			fullValue = refValue.Elem().Elem().Elem()
			fullType = refValue.Elem().Elem().Elem().Type()
		}
	}

	//logs.Debug("3 fullType ", fullType, " fullValue ", fullValue )

	tagName := "jsonp"
	jsonResult := make(map[string]interface{})

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)
		exists, tags := this.getTagsByTagName(field, tagName)

		attr := ""

		if !exists {
			continue
		}

		if len(tags) > 0 {
			attr = tags[0]
		}

		if len(strings.TrimSpace(attr)) == 0 {
			attr = Underscore(field.Name)
		}

		//logs.Debug("Field ", attr)

		fieldStruct := fullValue.FieldByName(field.Name)
		fieldValue := fieldStruct.Interface()

		if err := this.convertItem(jsonResult, attr, tags, fieldStruct, fieldValue); err != nil {
			return nil, err
		}

	}

	if writer, ok := obj.(JsonWriter); ok {
		writer.Write(jsonResult)
	}

	//logs.Debug("## filter tenant")
	return jsonResult, nil
}

func (this *JSON) convertItem(jsonResult map[string]interface{}, attr string, tags []string, fieldStruct reflect.Value, fieldValue any) error {
	ftype := fieldStruct.Type()
	isPtr := ftype.Kind() == reflect.Ptr
	isInterface := ftype.Kind() == reflect.Interface
	realKind := ftype.Kind()
	realType := ftype

	if reflect.TypeOf(fieldValue) == nil {
		return nil
	}

	// retorn true para &[]*Entity{}
	realTypePrt := false
	if isInterface {
		realTypePrt = reflect.TypeOf(fieldValue).Kind() == reflect.Ptr
	}

	if isPtr {
		realKind = ftype.Elem().Kind()
		realType = ftype.Elem()
	} else if isInterface && realTypePrt {
		realKind = reflect.TypeOf(fieldValue).Elem().Kind()
		realType = reflect.TypeOf(fieldValue).Elem()
	} else if isInterface {
		realKind = reflect.TypeOf(fieldValue).Kind()
		realType = reflect.TypeOf(fieldValue)
	}

	if this.Debug {
		//logs.Debug("Attr = ", attr, ", Field = ", field.Name, ", Type = ", ftype, "Kind = ", fieldStruct.Type().Kind(), ", Real Kind", realKind, "isPtr = ", isPtr) //, ", Value = ", fieldValue)
	}

	switch realKind {
	case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:

		jsonResult[attr] = fieldValue

		break

	case reflect.Slice:

		slice := reflect.ValueOf(fieldValue)
		//logs.Debug("slice 1 ", slice)
		zero := reflect.Zero(reflect.TypeOf(slice)).Interface() == slice

		if slice.IsNil() || zero {
			return nil
		}

		//logs.Debug("slice 2 ", slice)

		if isPtr || (isInterface && realTypePrt) {
			slice = slice.Elem()
		}

		//logs.Debug("slice", slice)

		sliceData := []interface{}{}
		for i := 0; i < slice.Len(); i++ {
			item := slice.Index(i)

			itype := reflect.TypeOf(item.Interface()).Kind()

			if itype == reflect.Ptr {
				itype = reflect.TypeOf(item.Interface()).Elem().Kind()
			}

			switch itype {
			case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:
				sliceData = append(sliceData, item.Interface())
				break
			case reflect.Struct:
				it, e := this.ToMap(item.Interface())
				if e != nil {
					return e
				}
				sliceData = append(sliceData, it)
				break
			default:
				logs.Debug("SLICE DATATYPE NOT FOUND: ", itype)
			}

		}

		//logs.Debug("sliceData = ", sliceData)
		jsonResult[attr] = &sliceData

		break

	case reflect.Map:

		jsonResult[attr] = fieldValue

	case reflect.Struct:

		zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue

		if zero {
			return nil
		}

		if realType.String() == timeStringKind {
			v, err := this.formatTime(fieldValue, isPtr, tags)

			if err != nil {
				return err
			}
			jsonResult[attr] = v
			break
		} else {

			if !isPtr {

				if fieldStruct.CanAddr() {
					addr := fieldStruct.Addr()
					fieldValue = addr.Interface()
				}

			}

			var e error
			//logs.Debug("to map ", reflect.TypeOf(fieldValue))
			jsonResult[attr], e = this.ToMap(fieldValue)
			if e != nil {
				return e
			}
		}
	}

	return nil
}

func (this *JSON) DecodeFromString(jsonStr string, obj interface{}) error {
	return this.Decode([]byte(jsonStr), obj)
}

func (this *JSON) Decode(b []byte, obj interface{}) error {

	dataMap := make(map[string]interface{})
	err := json.Unmarshal(b, &dataMap)
	if err != nil {
		return errors.New(fmt.Sprintf("json.Unmarshal: %v", err))
	}

	if this.DebugParse {
		logs.Debug("JSON = ", string(b))
	}

	return this.DecodeFromMap(dataMap, obj)
}

func (this *JSON) DecodeFromMap(jsonData map[string]interface{}, obj interface{}) error {

	defer func() {
		if r := recover(); r != nil {
			logs.Debug("DECODE FROM MAP ERROR: ", r, ", OBJ = ", obj)
		}
	}()

	if writer, ok := obj.(JsonWriter); ok {
		writer.Write(jsonData)
	}

	// value e type of pointer
	refValue := reflect.ValueOf(obj)
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	if refValue.Elem().Kind() == reflect.Interface {
		fullValue = refValue.Elem().Elem()
		fullType = refValue.Elem().Elem().Type()
	}

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)
		exists, tags := this.getTagsByTagName(field, tagName)
		attr := ""

		//logs.Debug("get value ", field.Name)
		if !exists {
			continue
		}

		if len(tags) > 0 {
			attr = tags[0]
		}

		if len(strings.TrimSpace(attr)) == 0 {
			attr = Underscore(field.Name)
		}

		if val, ok := jsonData[attr]; ok {

			fieldStruct := fullValue.FieldByName(field.Name)
			fieldValue := fieldStruct.Interface()

			ftype := fieldStruct.Type()
			isPtr := ftype.Kind() == reflect.Ptr
			realKind := ftype.Kind()
			realType := ftype
			if isPtr {
				realKind = ftype.Elem().Kind()
				realType = ftype.Elem()
			}

			value, err := this.getJsonValue(realType, jsonData, attr, tags)

			if err != nil {
				return err
			}

			if this.Debug {
				logs.Debug("Attr = ", attr, ", Field = ", field.Name, ", Type = ", ftype, "Kind = ", fieldStruct.Type().Kind(), ", Real Kind", realKind, ", Value = ", val, "isPtr = ", isPtr)
			}

			switch realKind {
			case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:
				//reflectValue := reflect.ValueOf(value)

				valueOf := reflect.ValueOf(value)
				reflectionValue := reflect.New(realType)
				converted := valueOf.Convert(realType)

				reflectionValue.Elem().Set(converted)

				if isPtr {
					reflectValue := reflectionValue.Interface()
					fieldStruct.Set(reflect.ValueOf(reflectValue))
				} else {
					reflectValue := reflectionValue.Elem().Interface()
					fieldStruct.Set(reflect.ValueOf(reflectValue))
				}
				break
			case reflect.Slice:

				reflection := reflect.MakeSlice(reflect.SliceOf(realType.Elem()), 0, 0)
				reflectionValue := reflect.New(reflection.Type())
				reflectionValue.Elem().Set(reflection)
				slice := reflectionValue.Interface()
				slicePtr := reflect.ValueOf(slice)
				sliceValuePtr := slicePtr.Elem()

				isItemPtr := realType.Elem().Kind() == reflect.Ptr
				itemRealType := realType.Elem()
				itemRealKind := realType.Elem().Kind()

				if isItemPtr {
					itemRealType = itemRealType.Elem()
					itemRealKind = itemRealType.Kind()
				}

				switch itemRealKind {
				case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:

					ds := value.([]interface{})

					for _, it := range ds {
						valueOf := reflect.ValueOf(it)
						realValue := valueOf.Convert(itemRealType)
						sliceValuePtr.Set(reflect.Append(sliceValuePtr, realValue))
					}

					break
				case reflect.Struct:

					ds := value.([]map[string]interface{})
					for _, it := range ds {
						newRefValue := reflect.New(itemRealType)
						newValue := newRefValue.Interface()
						this.DecodeFromMap(it, newValue)
						sliceValuePtr.Set(reflect.Append(sliceValuePtr, reflect.ValueOf(newValue)))
					}
					break
				}

				if !isPtr {
					value = slicePtr.Elem().Interface()
					reflectValue := reflect.ValueOf(value)
					fieldStruct.Set(reflectValue)
				} else {
					reflectValue := reflect.ValueOf(slicePtr.Interface())
					fieldStruct.Set(reflectValue)
				}

				break

			case reflect.Map:

				mapData := reflect.MakeMap(realType)
				//logs.Debug("mapData = ", mapData, " key ", realType.Key(), " elem ", realType.Elem())

				var reflectValue reflect.Value
				var mapRef interface{}
				mapValue := value.(map[string]interface{})

				mapValKind := realType.Elem().Kind()
				mapValType := realType.Elem()

				//mapKeyKind := realType.Key().Kind()
				mapKeyType := realType.Key()

				switch mapValKind {
				case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:

					for k, v := range mapValue {
						valueOfVal := reflect.ValueOf(v)
						convertedVal := valueOfVal.Convert(mapValType)

						valueOfKey := reflect.ValueOf(k)
						convertedKey := valueOfKey.Convert(mapKeyType)

						mapData.SetMapIndex(convertedKey, convertedVal)
					}

					mapRef = mapData
					value = mapData.Interface()
					break

				case reflect.Interface:
					mapRef = &mapValue
					break
				}

				if isPtr {
					reflectValue = reflect.ValueOf(mapRef)
				} else {
					reflectValue = reflect.ValueOf(value)
				}

				fieldStruct.Set(reflectValue)

				break

			case reflect.Struct:

				if realType.String() == timeStringKind {

					dVal := this.GetJsonString(jsonData, attr)
					if len(strings.TrimSpace(dVal)) > 0 {
						v, err := this.parseTime(isPtr, dVal, tags)

						if err != nil {
							return err
						}

						reflectValue := reflect.ValueOf(v)
						fieldStruct.Set(reflectValue)
					}

				} else {

					if data, ok := jsonData[attr].(map[string]interface{}); ok {

						var fieldFullType reflect.Type

						if isPtr {
							fieldFullType = reflect.TypeOf(fieldValue).Elem()
						} else {
							fieldFullType = reflect.TypeOf(fieldValue)
						}

						newRefValue := reflect.New(fieldFullType)
						fieldValue = newRefValue.Interface()

						this.DecodeFromMap(data, fieldValue)
						//logs.Debug("value of = ", fieldFullType, "%#v", fieldValue)
						if isPtr {
							fieldStruct.Set(newRefValue)
						} else {
							fieldStruct.Set(newRefValue.Elem())
						}

					} else {
						logs.Debug("Type not found to parse: field = ", field.Name, " type = ", fieldStruct.Type(), " value = ", val)
					}
				}
				break
			}
		}
	}
	//logs.Debug("## filter tenant")
	return nil
}

func (this *JSON) parseTime(ptr bool, s string, tags []string) (interface{}, error) {
	//s := this.GetJsonString(jsonData, attr)
	var value time.Time
	var err error
	var expectedFormat string = "unknow"

	if this.hasTagByName(tags, "date") {
		value, err = util.DateParse(this.DateFormat, s)
		expectedFormat = this.DateFormat
	} else if this.hasTagByName(tags, "datetime") {
		value, err = util.DateParse(this.DateTimeFormat, s)
		expectedFormat = this.DateTimeFormat
	} else if this.hasTagByName(tags, "time") {
		value, err = util.DateParse(this.TimeFormat, s)
		expectedFormat = this.TimeFormat
	} else {

		format, ok := this.hasFormatTag(tags)

		if ok {
			value, err = util.DateParse(format, s)
			expectedFormat = format
		} else {
			for _, layout := range this.DateLayouts {
				if len(layout) == len(s) {
					value, err = util.DateParse(this.TimestampFormat, s)

					if err != nil {
						break
					}
				}
			}
			expectedFormat = "???"
		}
	}

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error on parse time. Value: %v, Expected format: %v", s, expectedFormat))
	}

	if ptr {
		return &value, err
	}

	return value, err
}

func (this *JSON) formatTime(fieldValue interface{}, ptr bool, tags []string) (string, error) {

	var date time.Time

	if ptr {
		dt := fieldValue.(*time.Time)
		date = *dt
	} else {
		date = fieldValue.(time.Time)
	}

	var value string
	var err error
	var expectedFormat string = "unknow"

	if this.hasTagByName(tags, "date") {
		value = date.Format(this.DateFormat)
		expectedFormat = this.DateFormat
	} else if this.hasTagByName(tags, "datetime") {
		value = date.Format(this.DateTimeFormat)
		expectedFormat = this.DateTimeFormat
	} else if this.hasTagByName(tags, "time") {
		value = date.Format(this.TimeFormat)
		expectedFormat = this.TimeFormat
	} else {

		format, ok := this.hasFormatTag(tags)

		if ok {
			value = date.Format(format)
			expectedFormat = format
		} else {
			// "timestamp" is default
			value = date.Format(this.TimestampFormat)
			expectedFormat = this.TimestampFormat
		}
	}

	if err != nil {
		return "", errors.New(fmt.Sprintf("Error on format time.  Value: %v, Expected format: %v", date, expectedFormat))
	}
	return value, err
}

func (this *JSON) getJsonValue(rtype reflect.Type, jsonData map[string]interface{}, attr string, tags []string) (interface{}, error) {

	//timeType := reflect.TypeOf(time.Time{}).Kind()
	//timePtr := reflect.TypeOf(new(time.Time)).Kind()
	var value interface{}

	//logs.Debug("attr = ", attr,  " rtype.Kind() = ", rtype.Kind())

	switch rtype.Kind() {
	case reflect.Int64:
		value = this.GetJsonInt64(jsonData, attr)
		break
	case reflect.Int:
		value = this.GetJsonInt(jsonData, attr)
		break
	case reflect.Bool:
		value = this.GetJsonBool(jsonData, attr)
		break
	case reflect.Float32:
		value = this.GetJsonFloat32(jsonData, attr)
		break
	case reflect.Float64:
		value = this.GetJsonFloat64(jsonData, attr)
		break
	case reflect.String:
		value = this.GetJsonString(jsonData, attr)
		break
	case reflect.Slice:

		switch rtype.Elem().Kind() {
		case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:
			value = this.GetJsonSimpleArray(jsonData, attr)
		case reflect.Ptr:
			value = this.GetJsonArray(jsonData, attr)
		case reflect.Map:
			value = this.GetJsonObject(jsonData, attr)
		}

		break

	case reflect.Map:
		value = this.GetJsonObject(jsonData, attr)
		break

	default:

		isTime := false
		isPtr := rtype.Kind() == reflect.Ptr

		if rtype.String() == timeStringKind {
			isTime = true
		}

		if !isTime && isPtr {
			t := rtype.Elem()
			isTime = t.String() == timeStringKind
		}

		if isTime {
			dVal := this.GetJsonString(jsonData, attr)

			if len(strings.TrimSpace(dVal)) > 0 {
				v, err := this.parseTime(isPtr, dVal, tags)

				if err != nil {
					return nil, err
				}
				value = v
			} else {
				value = time.Time{}
			}
			break
		} else {

			if data, ok := jsonData[attr].(map[string]interface{}); ok {

				value = data

			} else {
				logs.Debug("Type not found to parse: attr = ", attr, " type = ", rtype)
			}

		}
		break
	}

	return value, nil
}

func (this *JSON) hasTagByName(tags []string, tagName string) bool {

	for _, tag := range tags {
		if tag == tagName {
			return true
		}
	}

	return false
}

func (this *JSON) hasFormatTag(tags []string) (string, bool) {

	for _, tag := range tags {
		if strings.Contains(tag, "format(") {
			format := strings.Replace(tag, "format(", "", -1)
			return format[0 : len(format)-1], true
		}
	}

	return "", false
}

func (this *JSON) getTagsByTagName(field reflect.StructField, tagName string) (bool, []string) {

	tag, exists := field.Tag.Lookup(tagName)
	var tags []string

	if len(strings.TrimSpace(tag)) > 0 {
		tags = strings.Split(tag, ";")
		return true, tags
	}

	return exists, tags
}

func Decode(b []byte, obj interface{}) error {
	return NewJSON().Decode(b, obj)
}

func Encode(obj interface{}) ([]byte, error) {
	return NewJSON().Encode(obj)
}

func EncodeToString(obj interface{}) (string, error) {
	return NewJSON().EncodeToString(obj)
}
