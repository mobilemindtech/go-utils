package response

import "fmt"

type ValidationMap struct {
	Messages map[string]string
}

func NewValidationMap() *ValidationMap {
	return &ValidationMap{Messages: map[string]string{}}
}

func ValidationMapOk() *ValidationMap {
	return NewValidationMap()
}

func (this *ValidationMap) AddMessage(key string, msg string, args ...interface{}) *ValidationMap {
	this.Messages[key] = fmt.Sprintf(msg, args...)
	return this
}

func (this *ValidationMap) IsOk() bool {
	return this.Empty()
}

func (this *ValidationMap) Count() int {
	return len(this.Messages)
}

func (this *ValidationMap) Empty() bool {
	return this.Count() == 0
}

func (this *ValidationMap) NonEmpty() bool {
	return !this.Empty()
}

func (this *ValidationMap) Error() string {
	return "validation error"
}
