package misc

import (
	"mime/multipart"
	"strings"
)

type Multipart struct {
	FileHeader *multipart.FileHeader
	File       *multipart.File
	Key        string
}

func (this *Multipart) FileName() string {
	return this.FileHeader.Filename
}

func (this *Multipart) FileExtension() string {
	ex := ""
	splited := strings.Split(this.FileName(), ".")

	if len(splited) > 0 {
		ex = splited[len(splited)-1]
	}
	return strings.ToLower(ex)
}
