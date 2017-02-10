package validator

import (  
  "github.com/mobilemindtec/go-utils/validator/cnpj"   
  "github.com/mobilemindtec/go-utils/validator/cpf"
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/astaxie/beego/validation"  
  "strings"
  _"fmt" 
)

func SetDefaultMessages() {
  validation.SetDefaultMessage(map[string]string{
    "Required":     "O valor para o campo é obrigatório",
    "Min":          "O valor máximo permitido é %d",
    "Max":          "O valor mínimo permitido é %d",
    "Range":        "Infome valores entre %d e %d",
    "MinSize":      "O tamanho mínimo para o texto é %d",
    "MaxSize":      "O tamanho máximo para o texto é %d",
    "Length":       "O tamanho necessário para o texto é %d",
    "Alpha":        "Apenas letras são permitidas",
    "Numeric":      "Apenas números são permitidos",
    "AlphaNumeric": "Apenas letras e números são permitidos",
    "Match":        "O valor para campo deve conferir com %s",
    "NoMatch":      "O valor para campo não comfere com %s",
    "AlphaDash":    "Must be valid alpha or numeric or dash(-_) characters",
    "Email":        "Informe um email válido",
    "IP":           "Informe um IP válido",
    "Base64":       "O valor para o campo deve ser Base64",
    "Mobile":       "Inform um celular válido",
    "Tel":          "Infoeme um telefone válido",
    "Phone":        "Informe um celular ou telefone válido",
    "ZipCode":      "Informe um cep válido",
    "Cnpj":         "O CNPJ informado é inválido",
    "Cpf":          "O CPF informado é inválido",
    "RequiredRel":  "Selecione uma opção",
  })

}

func AddCnpjValidator() {
  validation.AddCustomFunc("Cnpj", func(v *validation.Validation, obj interface{}, key string){ 

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
  validation.AddCustomFunc("Cpf", func(v *validation.Validation, obj interface{}, key string){ 

    key = strings.Split(key, ".")[0]

    s, _ := obj.(string)

    if len(strings.TrimSpace(s)) > 0 {
      if ok, _ := cpf.IsValid(s); !ok {
        v.SetError(key, validation.MessageTmpls["Cpf"])
      }
    }


  })
}

func AddRelationValidator() {

  validation.AddCustomFunc("RequiredRel", func(v *validation.Validation, obj interface{}, key string){ 


    key = strings.Split(key, ".")[0]

    isSatisfied := true

    if obj == nil {
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