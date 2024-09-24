package response

import "fmt"

type Validation struct {
	Messages []string
}

func NewValidation() *Validation {
	return &Validation{Messages: []string{}}
}

func ValidationOk() *Validation {
	return NewValidation()
}

func ValidationWith(msgs ...string) *Validation {
	return NewValidation().AddMessage(msgs...)
}

func (this *Validation) AddMessage(msgs ...string) *Validation {
	for _, msg := range msgs {
		this.Messages = append(this.Messages, msg)
	}
	return this
}

func (this *Validation) AddMsg(msg string, args ...interface{}) *Validation {
	this.Messages = append(this.Messages, fmt.Sprintf(msg, args...))
	return this
}

func (this *Validation) IsOk() bool {
	return this.Empty()
}

func (this *Validation) Count() int {
	return len(this.Messages)
}

func (this *Validation) Empty() bool {
	return this.Count() == 0
}

func (this *Validation) NonEmpty() bool {
	return !this.Empty()
}

func (this *Validation) Error() string {
	return "validation error"
}
