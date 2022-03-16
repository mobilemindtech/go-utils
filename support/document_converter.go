
package support

import(
  "encoding/base64"
  "compress/gzip"
  "strings"
  "errors"
  "bytes"
  "fmt"
  "os"
  "io"
)

type DocumentConverter struct{

}

func (this *DocumentConverter) SaveBase64Gzip(content string, path string) error{

  data, err := base64.StdEncoding.DecodeString(content)

  if err != nil {
    return err
  }

  bReader := bytes.NewReader(data)


  gReader, err := gzip.NewReader(bReader)

  if err != nil {
    return err
  }

  defer gReader.Close()

  writer, err := os.Create(path)

  if err != nil {
    return err
  }

  defer writer.Close()

  if _, err = io.Copy(writer, gReader); err != nil {
    return err
  }  

  return nil
}

func (this *DocumentConverter) GetExtension(documentName string) (string, error){

  splited := strings.Split(documentName, ".")

  if len(splited) == 0 {
    return "", errors.New(fmt.Sprintf("file extension not found: %v", documentName))
  }

  ext := splited[len(splited)-1]

  return ext, nil
	
}