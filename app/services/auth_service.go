package services

import (
	"github.com/mobilemindtec/go-utils/app/models"
	"strings"
	"errors"
)

type AuthService struct{
	UserInfo *models.User
}

func NewAuthService(user *models.User) *AuthService{
	return &AuthService{ UserInfo: user }
}

func (this *AuthService) IsAdmin() bool{
	if this.UserInfo != nil && this.UserInfo.Roles != nil {
		for _, it := range *this.UserInfo.Roles {
			if it.Authority == "ROLE_ADMIN" {
				return true
			}
		}
		return this.IsRoot()
	}
	return false
}

func (this *AuthService) IsRoot() bool{
	if this.UserInfo != nil && this.UserInfo.Roles != nil {
		for _, it := range *this.UserInfo.Roles {
			if it.Authority == "ROLE_ROOT" {
				return true
			}
		}
	}
	
	return false
}

func IsRootUser(user *models.User) bool{
	if user != nil && user.Roles != nil {
		for _, it := range *user.Roles {
			if it.Authority == "ROLE_ROOT" {
				return true
			}
		}
	}
	
	return false
}

func (this *AuthService) ValidPassword(password1 string, password2 string) error {

	if len(strings.TrimSpace(password1)) == 0 {
		return errors.New("A senha não pode ser vazia")
	}


	if password1 != password2 {
		return errors.New("As senhas devem ser iguais")
	}

	if len(password1) < 4 || len(password1) > 10 {
		return errors.New("O tamanho da senha deve ser de 4 a 10 caracteres")
	}

	return nil
}


func (this *AuthService) GetUserRoles() []string {

	roles := []string{}
	if this.UserInfo != nil && this.UserInfo.Roles != nil {
		for _, it := range *this.UserInfo.Roles {
			roles = append(roles, it.Authority)
		}
	}
	return roles
}