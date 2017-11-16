package services

import (
  "github.com/mobilemindtec/go-utils/app/models"  
  "github.com/mobilemindtec/go-utils/beego/db"
  "fmt"
)

type  AuditorInfo struct {
  Tenant *models.Tenant
  User *models.User  
}

type AuditorService struct {  
  Lang string  
  AuditorInfo *AuditorInfo
  Session *db.Session
} 

func NewAuditorService(session *db.Session, lang string, info *AuditorInfo) *AuditorService{
  return &AuditorService{ Lang: lang, AuditorInfo: info, Session: session }
}

func (this *AuditorService) OnAudit(format string, v ...interface{}) {
  content := fmt.Sprintf(format, v...)
  auditor := models.NewAuditorWithTenantAndContent(this.AuditorInfo.Tenant, content)
  auditor.User = this.AuditorInfo.User  
  
  if err := this.Session.Save(auditor); err != nil {
    fmt.Println("## error on save auditor: ", err.Error())
  }   
}

func (this *AuditorService) OnAuditWithNewDbSession(format string, v ...interface{}) {

  action := func() {
    content := fmt.Sprintf(format, v...)
    auditor := models.NewAuditorWithTenantAndContent(this.AuditorInfo.Tenant, content)
    auditor.User = this.AuditorInfo.User  
    
    session := db.NewSession()
    session.OpenWithoutTransaction()  
    defer session.Close()
    
    if err := session.Save(auditor); err != nil {
      fmt.Println("## error on save auditor: ", err.Error())
    }       
  }

  go action()

  
}