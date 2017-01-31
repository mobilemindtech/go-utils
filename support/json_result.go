package support

type JsonResult struct {
  Error bool
  Message string
  Results interface{}
  Result interface{}

  Errors map[string]string

  CurrentUnixTime int64
}