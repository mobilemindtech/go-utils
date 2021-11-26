package json

import(
	"github.com/mobilemindtec/go-utils/support"
	"github.com/mobilemindtec/go-utils/app/util"
	"encoding/json"
	"reflect"
	"strings"
	"errors"
	"time"
	"fmt"
)

const (
	TimestampLayout string = "2006-01-02T15:04:05-07:00"
	DateLayout string = "2006-02-01"
  DateTimeLayout = "2006-02-01 15:04:05"
  TimeLayout = "10:25:05"	
  timeStringKind = "time.Time"
  tagName = "jsonp"

)

type JSON struct {
	support.JsonParser
	Debug bool
	DateFormat string
	DateTimeFormat string
	TimeFormat string
	TimestampFormat string

	DebugParse bool
	DebugFormat bool
}

func NewJSON() *JSON {
	return &JSON{ 
		DateFormat: DateLayout,
		DateTimeFormat: DateTimeLayout,
		TimeFormat: TimeLayout,
		TimestampFormat	:TimestampLayout,	
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

	data, err := this.ToMap(obj)

	if err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(data, "", " ")

	if this.DebugFormat {
		fmt.Println("JSON = ", string(result))
	}

	return result, err
}

func (this *JSON) ToMap(obj interface{}) (map[string]interface{}, error) {
  // value e type of pointer

  defer func() {
    if r := recover(); r != nil {
      fmt.Println("JSON TO MAP ERROR: ", r)
      panic(r)
    }
  }()

  refValue := reflect.ValueOf(obj)
  fullValue := refValue
  fullType := fullValue.Type()
  
  //fmt.Println("fullType ", fullType, " fullValue ", fullValue)

  if reflect.TypeOf(obj).Kind() == reflect.Ptr {
  	fullValue = refValue.Elem()
  	fullType = refValue.Elem().Type()

  }

  if fullValue.Kind() == reflect.Interface {
  	fullValue = refValue.Elem().Elem()
  	fullType = refValue.Elem().Elem().Type()  	
  }

  //fmt.Println("fullType ", fullType, " fullValue ", fullValue )

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

    //fmt.Println("Field ", attr)

    
    fieldStruct := fullValue.FieldByName(field.Name)
    fieldValue := fieldStruct.Interface()
  	
    ftype := fieldStruct.Type()
    isPtr := ftype.Kind() == reflect.Ptr
    isInterface := ftype.Kind() == reflect.Interface
    realKind := ftype.Kind()
    realType := ftype

		if reflect.TypeOf(fieldValue) == nil {
			continue
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

    //fmt.Println("real type ", realType, "real kind ", realKind, "is ptr ", isPtr, "is interface ", isInterface, "is real type ptr", realTypePrt)

   /* if isInterface {

    	//fmt.Println("is interface, ", attr, reflect.TypeOf(fieldValue))

    	if fieldValue == nil {
    		continue
    	}

    	realKind = reflect.TypeOf(fieldValue).Kind()
    	realType = reflect.TypeOf(fieldValue)


    	if realKind == reflect.Ptr {
	    	tValue := reflect.TypeOf(fieldValue).Elem()
	    	//fmt.Println("is interface, ", attr, reflect.TypeOf(fieldValue))

	  		realKind = reflect.TypeOf(tValue).Kind()
	  		realType = reflect.TypeOf(tValue)
	    	isPtr = realKind == reflect.Ptr
	    	//fmt.Println("is ptr, ", attr, isPtr)
	    	if isPtr {
	    		fieldValue = reflect.ValueOf(fieldValue).Interface()
	    		realKind = realType.Elem().Kind()
	    		realType = realType.Elem()
	    	}
    	} else {
    		fieldValue = reflect.ValueOf(fieldValue)
    	}
    }*/

  	if this.Debug {
  		fmt.Println("Attr = ", attr, ", Field = ", field.Name, ", Type = ", ftype , "Kind = ", fieldStruct.Type().Kind(), ", Real Kind", realKind, "isPtr = ", isPtr) //, ", Value = ", fieldValue)
  	}
  	
    
    switch realKind {
    	case reflect.Int64, reflect.Int, reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:

    		jsonResult[attr] = fieldValue

    		break	    		

    	case reflect.Slice:

    		
    		slice := reflect.ValueOf(fieldValue)
    		//fmt.Println("slice 1 ", slice)
    		zero := reflect.Zero(reflect.TypeOf(slice)).Interface() == slice

    		if slice.IsNil() || zero {
    			continue
    		}
    		
    		//fmt.Println("slice 2 ", slice)

    		if isPtr || (isInterface && realTypePrt) {
					slice = slice.Elem()    			
    		} 

    		//fmt.Println("slice", slice)


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
		    				return nil, e
		    			}
		    			sliceData = append(sliceData, it)
		    			break
		    		default:
		    			fmt.Println("SLICE DATATYPE NOT FOUND: ", itype)
    			}

    		}

    		//fmt.Println("sliceData = ", sliceData)
    		jsonResult[attr] = &sliceData

    		break

    	case reflect.Map:

    		jsonResult[attr] = fieldValue
    	
    	case reflect.Struct:

		  	zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue

		  	if zero {
		  		continue
		  	}


    		if realType.String() == timeStringKind {
	    		v, err := this.formatTime(fieldValue, isPtr, tags)

	    		if err != nil {
	    			return nil, err
	    		}
	    		jsonResult[attr] = v
	    		break	    			
    		} else {

    			if !isPtr {
    				fieldValue = fieldStruct.Addr().Interface()
    			}
    			

  				var e error
  				//fmt.Println("to map ", reflect.TypeOf(fieldValue))
  				jsonResult[attr], e = this.ToMap(fieldValue)
  				if e != nil {
  					return nil, e
  				}

    		}
    }      
	}
	//fmt.Println("## filter tenant")
	return jsonResult, nil	
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
		fmt.Println("JSON = ", string(b))
	}

	return this.DecodeFromMap(dataMap, obj)
}

func (this *JSON) DecodeFromMap(jsonData map[string]interface{}, obj interface{}) error{

  defer func() {
    if r := recover(); r != nil {
      fmt.Println("DECODE FROM MAP ERROR: ", r)
    }
  }()

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

	   //fmt.Println("get value ", field.Name)
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
    		fmt.Println("Attr = ", attr, ", Field = ", field.Name, ", Type = ", ftype , "Kind = ", fieldStruct.Type().Kind(), ", Real Kind", realKind, ", Value = ", val, "isPtr = ", isPtr)
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
	    		//fmt.Println("mapData = ", mapData, " key ", realType.Key(), " elem ", realType.Elem())

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
		    		v, err := this.parseTime(isPtr, this.GetJsonString(jsonData, attr), tags)

		    		if err != nil {
		    			return err
		    		}

				    reflectValue := reflect.ValueOf(v)
				    fieldStruct.Set(reflectValue)

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
			        //fmt.Println("value of = ", fieldFullType, "%#v", fieldValue)
			        if isPtr {
			        	fieldStruct.Set(newRefValue)
			        } else {
			        	fieldStruct.Set(newRefValue.Elem())
			        }
				     
			    	} else {
			    		fmt.Println("Type not found to parse: field = ", field.Name, " type = ", fieldStruct.Type(), " value = ", val)		    			
		    		}

	    		}
	    		break
	    }

    }   
	}
	//fmt.Println("## filter tenant")
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
	}	else {
		// timestamp is default
		value, err = util.DateParse(this.TimestampFormat, s)
		expectedFormat = this.TimestampFormat
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
		// "timestamp" is default
		value = date.Format(this.TimestampFormat)
		expectedFormat = this.TimestampFormat	
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

	//fmt.Println("attr = ", attr,  " rtype.Kind() = ", rtype.Kind())

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
    		v, err := this.parseTime(isPtr, this.GetJsonString(jsonData, attr), tags)

    		if err != nil {
    			return nil, err
    		}
    		value = v
    		break	    			
  		} else {

    		if data, ok := jsonData[attr].(map[string]interface{}); ok {

	    		value = data

	    	} else {
	    		fmt.Println("Type not found to parse: attr = ", attr, " type = ", rtype)		    			
    		}

  		}
  		break
  }	

  return value, nil
}


func (this *JSON) hasTagByName(tags []string, tagName string) bool{

  for _, tag := range tags {
    if tag == tagName {
      return true
    }
  }

  return false
}

func (this *JSON) getTagsByTagName(field reflect.StructField, tagName string) (bool, []string){

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