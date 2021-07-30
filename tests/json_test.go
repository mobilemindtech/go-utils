package goutils

import (
  "testing"
  "github.com/mobilemindtec/go-utils/json"
)

type PersonType int64

const (
  PersonTypeM PersonType = 1
  PersonTypeF PersonType = 2
)

type Age struct {
  Age int64 `jsonp:"age"`
}

type Person struct {
  Id int64 `jsonp:"id"`
  Name string `jsonp:"name"`
  Age *Age `jsonp:"age"`
  Now *time.Time `jsonp:";timestamp"`

  Friends []*Person `jsonp:""`
  Tags []string `jsonp:""`
  Data map[string]interface{} `jsonp:""`

  Now2 time.Time `jsonp:";timestamp"`
  Age2 Age `jsonp:"age2"`
  Friends2 *[]*Person `jsonp:""`
  Tags2 *[]string `jsonp:""`
  Data2 *map[string]interface{} `jsonp:""`
  Type PersonType `jsonp:""`
  N *int `jsonp:""`
}



func TestJson(t *testing.T) {

  JSON := json.NewJSON()
  json.Debug = true

  data := make(map[string]interface{})
  age := make(map[string]interface{})
  age["age"] = 34
  data["name"] = "Ricardo"
  data["id"] = 1
  data["age"] = age
  data["now"] = "2021-07-12T14:44:00-03:00"

  //p := new(Person)
  //json.ParseMap(p, data)
  //fmt.Println("PARSE MAP:", fmt.Sprintf("%#v, %#v", p, p.Age))

  pp := new(Person)
  JSON.Parse(pp, []byte(`
    {
      "id": 1,
      "name": "Jons",
      "age": {
        "age": 45
      },
      "age2": {
        "age": 30
      },
      "type": 1,
      "n": 5,
      "now": "2021-07-12T14:44:00-03:00",
      "now2": "2021-07-12T14:44:00-03:00",
      "tags": ["a", "b", "c"],
      "tags2": ["a", "b", "c"],
      "friends": [{"id": 2, "name": "Mark"}, {"id": 3, "name": "Juca"}],
      "friends2": [{"id": 2, "name": "Mark"}, {"id": 3, "name": "Juca"}],
      "data": {"x": 1, "y": 2},
      "data2": {"x": 1, "y": 2}
    }
  `))

  fmt.Println("--------------------------------------")
  fmt.Println("PARSE:", fmt.Sprintf("%#v", pp))
  fmt.Println("--------------------------------------")
  fmt.Println("Friends:", fmt.Sprintf("%#v", pp.Friends))
  fmt.Println("--------------------------------------")
  fmt.Println("Friends2:", fmt.Sprintf("%#v", pp.Friends2))
  fmt.Println("--------------------------------------")
  fmt.Println("Age:", fmt.Sprintf("%#v", pp.Age))
  fmt.Println("--------------------------------------")
  fmt.Println("Age2:", fmt.Sprintf("%#v", pp.Age2))
  fmt.Println("--------------------------------------")
  fmt.Println("Tags", fmt.Sprintf("%#v", pp.Tags))
  fmt.Println("--------------------------------------")
  fmt.Println("Tags2", fmt.Sprintf("%#v", pp.Tags2))
  fmt.Println("--------------------------------------")
  fmt.Println("Data", fmt.Sprintf("%#v", pp.Data))
  fmt.Println("--------------------------------------")
  fmt.Println("Data2", fmt.Sprintf("%#v", pp.Data2))
  fmt.Println("--------------------------------------")
  fmt.Println("Now", fmt.Sprintf("%#v", pp.Now))
  fmt.Println("--------------------------------------")
  fmt.Println("Now2", fmt.Sprintf("%#v", pp.Now2))
  fmt.Println("--------------------------------------")
  fmt.Println("N", *pp.N)

  //jsonData, err := json.ToMap(p)
  //fmt.Println(fmt.Sprintf("TO MAP: %#v, Err %v", jsonData, err))

  //b, err := json.Format(p)
  //fmt.Println(fmt.Sprintf("FORMAT: %#v, Err %v", string(b), err))


}