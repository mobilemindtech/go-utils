package services

type LoginErrorUserNotFound struct {
	Message string
}

func LoginUserNotFound(msg string) *LoginErrorUserNotFound {
	return &LoginErrorUserNotFound{msg}
}

func (e *LoginErrorUserNotFound) Error() string {
	return e.Message
}

type LoginErrorWrongPassword struct {
	Message string
}

func LoginWrongPassword(msg string) *LoginErrorWrongPassword {
	return &LoginErrorWrongPassword{msg}
}

func (e *LoginErrorWrongPassword) Error() string {
	return e.Message
}

type LoginErrorUserInactive struct {
	Message string
}

func LoginUserInactive(msg string) *LoginErrorUserInactive {
	return &LoginErrorUserInactive{msg}
}

func (e *LoginErrorUserInactive) Error() string {
	return e.Message
}

type LoginErrorTenantNotFound struct {
	Message string
}

func LoginTenantNotFound(msg string) *LoginErrorTenantNotFound {
	return &LoginErrorTenantNotFound{msg}
}

func (e *LoginErrorTenantNotFound) Error() string {
	return e.Message
}
