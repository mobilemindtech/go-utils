package resps

type HttpResponse interface {
	StatusCode() int
	Body() interface{}
}

type Response struct {
	statusCode int
	body       interface{}
	raw        bool
}

func (this *Response) StatusCode() int {
	return this.statusCode
}

func (this *Response) Body() interface{} {
	return this.body
}

func NewResponse(statusCode int, body interface{}) *Response {
	return &Response{statusCode: statusCode, body: body}
}

type HttpError interface {
	StatusCode() int
	Error() string
}

type Ok struct {
	body interface{}
}

func (this *Ok) StatusCode() int {
	return 200
}

func (this *Ok) Body() interface{} {
	return this.body
}

func NewOk(body interface{}) *Ok {
	return &Ok{body}
}

type Found struct {
	location string
}

func (this *Found) StatusCode() int {
	return 302
}

func (this *Found) Body() interface{} {
	return ""
}

func NewFound(location string) *Found {
	return &Found{location}
}

type NotFound struct {
	Message string
}

func (this *NotFound) Error() string {
	return this.Message
}
func (this *NotFound) Body() string {
	return this.Message
}
func (this *NotFound) StatusCode() int {
	return 404
}
func NewNotFound(msg ...string) *NotFound {
	message := "Not Found"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &NotFound{message}
}

type ServerError struct {
	Message string
}

func (this *ServerError) Error() string {
	return this.Message
}
func (this *ServerError) Body() string {
	return this.Message
}
func (this *ServerError) StatusCode() int {
	return 500
}
func NewServerError(msg ...string) *ServerError {
	message := "Server Error"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &ServerError{message}
}

type BadRequest struct {
	Message string
	Errors  []string
}

func (this *BadRequest) Error() string {
	return this.Message
}
func (this *BadRequest) Body() string {
	return this.Message
}
func (this *BadRequest) StatusCode() int {
	return 400
}
func (this *BadRequest) WithErrors(errors []string) *BadRequest {
	this.Errors = errors
	return this
}

func NewBadRequest(msg ...string) *BadRequest {
	message := "Bad Request"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &BadRequest{message, nil}
}

type Unauthorized struct {
	Message string
}

func (this *Unauthorized) Error() string {
	return this.Message
}
func (this *Unauthorized) Boyd() string {
	return this.Message
}
func (this *Unauthorized) StatusCode() int {
	return 401
}
func NewUnauthorized(msg ...string) *Unauthorized {
	message := "Unauthorized"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &Unauthorized{message}
}

type Forbidden struct {
	Message string
}

func (this *Forbidden) Error() string {
	return this.Message
}
func (this *Forbidden) Body() string {
	return this.Message
}
func (this *Forbidden) StatusCode() int {
	return 403
}
func NewForbidden(msg ...string) *Forbidden {
	message := "Forbidden"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &Forbidden{message}
}
