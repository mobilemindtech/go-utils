package fault

type UserNotFoundError struct {
	Message string
}

func (this *UserNotFoundError) Error() string {
	return "User not found"
}

func NewUserNotFoundError(msg ...string) *UserNotFoundError {
	message := "Unauthorized"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &UserNotFoundError{message}
}

type UserUnauthorizedError struct {
	Message string
}

func (this *UserUnauthorizedError) Error() string {
	return this.Message
}

func NewUserUnauthorizedError(msg ...string) *UserUnauthorizedError {
	message := "Forbidden"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &UserUnauthorizedError{message}
}

type ValidationError struct {
	Errors []string
	Message string
}

func (this *ValidationError) Error() string {
	return this.Message
}

func NewValidationError(errors ...[]string) *ValidationError {
	var errs []string
	if len(errors) > 0 {
		errs = errors[0]
	}
	return &ValidationError{errs, "Validation error"}
}

func (this *ValidationError) WithMsg(msg string) *ValidationError {
	this.Message = msg
	return this
}

type EntityNotFound struct {
	Message string
}

func (this *EntityNotFound) Error() string {
	return this.Message
}

func NewEntityNotFound(msg ...string) *EntityNotFound {
	message := "entity not found"
	if len(msg) > 0 {
		message = msg[0]
	}
	return &EntityNotFound{message}
}
