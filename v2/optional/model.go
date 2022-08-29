package optional

type Optional struct {
	
} 

type None struct {
	
}

func NewNone() *None {
	return &None{}
}

type Some struct {	
	Item interface{}
}

func NewSome(item interface{}) *Some {
	return &Some{Item: item}
}

func NewSomeEmpty() *Some {
	return &Some{}
}

type Try struct {
	
}

type Fail struct {
	Error error
}

func (this *Fail) ErrorString() string {
	return this.Error.Error()
}


func NewFail(err error) *Fail {
	return &Fail{ Error: err }
}

type Success struct {
	Item interface{}
}

func (this *Success) WithItem(item interface{}) *Success{
	this.Item = item
	return this
}

func NewSuccess() *Success {
	return &Success{ }
}

type Either struct {
	
}

type Left struct {
		Item interface{}
}

func (this *Left) WithItem(item interface{}) *Left{
	this.Item = item
	return this
}

func NewLeft() *Left{
	return &Left{}
}

type Rigth struct {
	Item interface{}
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

func GetOrDefault[R any](val interface{}, r R) R{
	if x, ok := val.(R); ok {
		return x
	}
	return r
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
			return x	
	}
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

func GetFailError(val interface{}) error{
	return val.(*Fail).Error
}
