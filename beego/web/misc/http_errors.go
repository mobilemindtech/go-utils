package misc

type HttpError interface {
	StatusCode() int
	Error() string
}

type NotFound struct {
	Message string
	Code    int
}

func (this *NotFound) Error() string {
	return this.Message
}
func (this *NotFound) StatusCode() int {
	return this.Code
}
func MakeNotFound() *NotFound {
	return &NotFound{"Not Found", 404}
}

type ServerError struct {
	Message string
	Code    int
}

func (this *ServerError) Error() string {
	return this.Message
}
func (this *ServerError) StatusCode() int {
	return this.Code
}
func MakeServerError() *ServerError {
	return &ServerError{"Server Error", 500}
}

type BadRequest struct {
	Message string
	Code    int
	Errors  []string
}

func (this *BadRequest) Error() string {
	return this.Message
}
func (this *BadRequest) StatusCode() int {
	return this.Code
}
func MakeBadRequest(errors ...[]string) *BadRequest {
	var errs []string
	if len(errors) > 0 {
		errs = errors[0]
	}
	return &BadRequest{"Bad Request", 400, errs}
}

type Unauthorized struct {
	Message string
	Code    int
}

func (this *Unauthorized) Error() string {
	return this.Message
}
func (this *Unauthorized) StatusCode() int {
	return this.Code
}
func MakeUnauthorized() *Unauthorized {
	return &Unauthorized{"Unauthorized", 401}
}

type Forbidden struct {
	Message string
	Code    int
}

func (this *Forbidden) Error() string {
	return this.Message
}
func (this *Forbidden) StatusCode() int {
	return this.Code
}
func MakeForbidden() *Forbidden {
	return &Forbidden{"Forbidden", 401}
}
