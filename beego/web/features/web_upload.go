package features

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/beego/web/misc"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/v2/optional"
	"github.com/satori/go.uuid"
)

type WebUpload struct {
	UploadPathDestination string
	base                  trait.WebBaseInterface
}

func (this *WebUpload) InitWebUpload(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebUpload) PrepareUploadedFile(fileOriginalName string, fileName string) (string, string, error) {

	path := this.UploadPathDestination + string(os.PathSeparator)

	if err := os.MkdirAll(path, 0777); err != nil {
		return "", "", err
	}

	splited := strings.Split(fileOriginalName, ".")

	if len(splited) == 0 {
		return "", "", errors.New("file extension not found")
	}

	ext := splited[len(splited)-1]

	if !strings.Contains(fileName, ".") {
		fileName += "." + ext
	}

	path += fileName

	logs.Trace("## save file on ", path)

	return path, fileName, nil
}

func (this *WebUpload) GetUploadedFileSavePath(fieldName string) string {

	if err := os.MkdirAll(this.UploadPathDestination, 0777); err != nil {
		logs.Error("Error on create uploaded file save path %v: %v", this.UploadPathDestination, err)
	}

	return fmt.Sprintf("%v/%v", this.UploadPathDestination, fieldName)
}

func (this *WebUpload) HasUploadedFile(fname string) (bool, error) {
	_, _, err := this.base.GetBeegoController().GetFile(fname)
	if err == http.ErrMissingFile {
		return false, nil
	} else if err != nil {
		return false, errors.New(fmt.Sprintf("Erro ao buscar documento: %v", err))
	}
	return true, nil
}

func (this *WebUpload) GetUploadedFileExt(fieldName string, required bool) (bool, string, error) {

	_, multipartFileHeader, err := this.base.GetBeegoController().GetFile(fieldName)

	if err == http.ErrMissingFile {

		if required {
			return false, "", fmt.Errorf("Selecione uma imagem para enviar")
		}

		return false, "", nil

	} else if err != nil {
		return false, "", fmt.Errorf("Erro ao enviar arquivo: %v", err)
	}

	originalName := this.base.GetCharacter().Transform(multipartFileHeader.Filename)
	splited := strings.Split(originalName, ".")

	if len(splited) == 0 {
		return false, "", fmt.Errorf("file extension not found")
	}

	ext := splited[len(splited)-1]
	uuid := uuid.NewV4()
	newFileName := fmt.Sprintf("%v.%v", uuid.String(), ext)

	return true, newFileName, nil
}

func (this *WebUpload) GetFileOpt(key string) *optional.Optional[*misc.Multipart] {
	file, fileHeader, err := this.base.GetBeegoController().GetFile(key)

	if err != nil {
		if err == http.ErrMissingFile {
			return optional.OfNone[*misc.Multipart]()
		}
		return optional.OfFail[*misc.Multipart](err)
	}

	return optional.Of[*misc.Multipart](&misc.Multipart{File: &file, FileHeader: fileHeader, Key: key})
}
