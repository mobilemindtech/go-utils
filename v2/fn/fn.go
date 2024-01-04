package fn

import (
	_ "github.com/mobilemindtec/go-utils/v2/optional"
	"reflect"
)


type FuncInfo struct {
	Type reflect.Type
	ArgsCount int
	Value reflect.Value
	ArgsTypes []reflect.Type
}

func (this *FuncInfo) ArgType(i int) reflect.Type {
	return this.ArgsTypes[i]
}

func (this *FuncInfo) Call(vals []reflect.Value) []reflect.Value {
	return this.Value.Call(vals)
}

func (this *FuncInfo) CallEmpty() []reflect.Value {
	return this.Call([]reflect.Value{})
}

func (this *FuncInfo) HasTypedArgs() bool {
	return len(this.ArgsTypes) > 0
}

func NewFuncInfo(f interface{}) *FuncInfo {
	fnType := reflect.TypeOf(f)
	fnArgsCount := fnType.NumIn()
	fnValue := reflect.ValueOf(f)
	argsTypes := []reflect.Type{}
	for i := 0; i < fnArgsCount; i++ {
		argType := fnType.In(i)
		if argType != nil && argType.Kind() != reflect.Interface { // typeof == nil is arg interface
			argsTypes = append(argsTypes, argType)
		}
	}

	return  &FuncInfo{
		Type: fnType,
		ArgsCount: fnArgsCount,
		Value: fnValue,
		ArgsTypes: argsTypes,
	}
}
