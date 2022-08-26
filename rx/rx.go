package rx

type Optional interface {
	
} 

type None struct {
	Optional
}

func NewNone() *None {
	return &None{}
}

type Some struct {
	Optional
	Item interface{}
}

func NewSome(item interface{}) *Some {
	return &Some{Item: item}
}

func NewSomeEmpty() *Some {
	return &Some{}
}


//func (this *Some) GetOrDefault(val T) T {
//	if  this.Item != *new(T) {
//		return this.Item
//	}
//	return val
//}

type Try interface {
	Optional
}

type Fail struct {
	Try
	Error error
}

func (this *Fail) ErrorString() string {
	return this.Error.Error()
}


func NewFail(err error) *Fail {
	return &Fail{ Error: err }
}

type Success struct {
	Try
	Item interface{}
}

func (this *Success) WithItem(item interface{}) *Success{
	this.Item = item
	return this
}

func NewSuccess() *Success {
	return &Success{ }
}

type Left struct {
	Success
}

func (this *Left) WithItem(item interface{}) *Left{
	this.Item = item
	return this
}


func NewLeft() *Left{
	return &Left{}
}

type Rigth struct {
	Success
}

func (this *Rigth) WithItem(item interface{}) *Rigth{
	this.Item = item
	return this
}

func NewRigth() *Rigth{
	return &Rigth{}
}

func Get[R any](val interface{}) R{
	return val.(R)
}

func GetItem[R any](val interface{}) R{
	switch val.(type) {
		case Some:
			return GetSome(val).Item.(R)
		case Success:
			return GetSuccess(val).Item.(R)
		case Left:
			return GetLeft(val).Item.(R)
		case Rigth:
			return GetRigth(val).Item.(R)
		default: 
			var x R
			return x	}
}

func GetFail(val interface{}) *Fail{
	return val.(*Fail)
}

func GetSuccess(val interface{}) *Success{
	return val.(*Success)
}

func GetSome(val interface{}) *Some{
	return val.(*Some)
}

func GetLeft(val interface{}) *Left{
	return val.(*Left)
}

func GetRigth(val interface{}) *Rigth{
	return val.(*Rigth)
}