package validator

import (
	_ "fmt"
	"github.com/beego/beego/v2/core/validation"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/validator/cnpj"
	"github.com/mobilemindtech/go-utils/validator/cpf"
	"reflect"
	"strings"
)

func SetDefaultMessages() {
	validation.SetDefaultMessage(map[string]string{
		"Required":      "O valor para o campo é obrigatório",
		"Min":           "O valor máximo permitido é %d",
		"Max":           "O valor mínimo permitido é %d",
		"Range":         "Infome valores entre %d e %d",
		"MinSize":       "O tamanho mínimo para o texto é %d",
		"MaxSize":       "O tamanho máximo para o texto é %d",
		"Length":        "O tamanho necessário para o texto é %d",
		"Alpha":         "Apenas letras são permitidas",
		"Numeric":       "Apenas números são permitidos",
		"AlphaNumeric":  "Apenas letras e números são permitidos",
		"Match":         "O valor para campo deve conferir com %s",
		"NoMatch":       "O valor para campo não comfere com %s",
		"AlphaDash":     "Must be valid alpha or numeric or dash(-_) characters",
		"Email":         "Informe um email válido",
		"IP":            "Informe um IP válido",
		"Base64":        "O valor para o campo deve ser Base64",
		"Mobile":        "Inform um celular válido",
		"Tel":           "Infoeme um telefone válido",
		"Phone":         "Informe um celular ou telefone válido",
		"ZipCode":       "Informe um cep válido",
		"Cnpj":          "O CNPJ informado é inválido",
		"Cpf":           "O CPF informado é inválido",
		"RequiredRel":   "Selecione uma opção",
		"RequiredConst": "Selecione uma opção",
	})

}

func AddCnpjValidator() {
	validation.AddCustomFunc("Cnpj", func(v *validation.Validation, obj interface{}, key string) {

		key = strings.Split(key, ".")[0]

		s, _ := obj.(string)

		if len(strings.TrimSpace(s)) > 0 {
			if ok, _ := cnpj.IsValid(s); !ok {
				v.SetError(key, validation.MessageTmpls["Cnpj"])
			}
		}

	})
}

func AddCpfValidator() {
	validation.AddCustomFunc("Cpf", func(v *validation.Validation, obj interface{}, key string) {

		key = strings.Split(key, ".")[0]

		s, _ := obj.(string)

		if len(strings.TrimSpace(s)) > 0 {
			if ok, _ := cpf.IsValid(s); !ok {
				v.SetError(key, validation.MessageTmpls["Cpf"])
			}
		}

	})
}

func AddConstValidator() {

	validation.AddCustomFunc("RequiredConst", func(v *validation.Validation, obj interface{}, key string) {

		key = strings.Split(key, ".")[0]

		ref := reflect.ValueOf(obj)

		if ref.Kind() != reflect.Int32 || ref.Kind() != reflect.Int64 {
			val := int(ref.Int())

			if int(val) <= 0 {
				v.SetError(key, validation.MessageTmpls["RequiredConst"])
			}

		} else {
			v.SetError(key, validation.MessageTmpls["RequiredConst"])
		}

	})
}

func AddRelationValidator() {

	validation.AddCustomFunc("RequiredRel", func(v *validation.Validation, obj interface{}, key string) {

		key = strings.Split(key, ".")[0]

		isSatisfied := true

		if support.IsNil(obj) {
			isSatisfied = false
		} else {
			if model, ok := obj.(db.Model); ok {
				if !model.IsPersisted() {
					isSatisfied = false
				}
			} else {
				v.SetError(key, "relation does not implements db.Model")
			}
		}

		if !isSatisfied {
			v.SetError(key, validation.MessageTmpls["RequiredRel"])
		}

	})
}
