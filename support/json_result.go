package support

type JsonResult struct {
  Error bool
  Message string
  Results interface{}
  Result interface{}

  Errors map[string]string

  TotalCount int64

  CurrentUnixTime int64
}