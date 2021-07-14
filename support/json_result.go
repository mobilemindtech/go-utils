package support

type JsonResult struct {
  Error bool `jsonp:""`
  Message string `jsonp:""`
  Results interface{} `jsonp:""`
  Result interface{} `jsonp:""`

  Errors map[string]string `jsonp:""`

  TotalCount int64 `jsonp:""`

  CurrentUnixTime int64 `jsonp:""`
}