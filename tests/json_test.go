package tests

import (
  "testing"
  "github.com/mobilemindtec/go-utils/json"
  "fmt"
  "time"
)

type PersonType int64
type PersonType2 string

const (
  PersonTypeM PersonType = 1
  PersonTypeF PersonType = 2
)

const (
  PersonType2M PersonType2 = "Masculino"
  PersonType2F PersonType2 = "Feminino"
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
  Data3 map[string]string `jsonp:""`

  Type2 PersonType2 `jsonp:""`

  Now2 time.Time `jsonp:";timestamp"`
  Age2 Age `jsonp:"age2"`
  Friends2 *[]*Person `jsonp:""`
  Tags2 *[]string `jsonp:""`
  Data2 *map[string]interface{} `jsonp:""`
  Type PersonType `jsonp:""`
  Types []PersonType2 `jsonp:""`
  Types2 *[]PersonType2 `jsonp:""`
  N *int `jsonp:""`

  Result interface{} `jsonp:""`
}

type Result struct {
  Data interface{} `jsonp:""`
}




// go test -v  github.com/mobilemindtec/go-utils/tests -run TestJson 
func TestJson(t *testing.T) {
  json := json.NewJSON()
  json.Debug = true

  data := make(map[string]interface{})
  age := make(map[string]interface{})
  age["age"] = 34
  data["name"] = "Ricardo"
  data["id"] = 1
  data["age"] = age
  data["now"] = "2021-07-12T14:44:00-03:00"

  //p := new(Person)
  //json.ParseMap(data, p)
  //fmt.Println("PARSE MAP:", fmt.Sprintf("%#v, %#v", p, p.Age))

  pp := new(Person)
  err := json.Decode([]byte(`
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
      "type2": "Masculino",
      "types": ["Masculino", "Feminino"],
      "types2": ["Masculino", "Feminino"],
      "n": 5,
      "now": "2021-07-12T14:44:00-03:00",
      "now2": "2021-07-12T14:44:00-03:00",
      "tags": ["a", "b", "c"],
      "tags2": ["a", "b", "c"],
      "friends": [{"id": 2, "name": "Mark"}, {"id": 3, "name": "Juca"}],
      "friends2": [{"id": 2, "name": "Mark"}, {"id": 3, "name": "Juca"}],
      "data": {"x": 1, "y": 2},
      "data2": {"x": 1, "y": 2},
      "data3": {"x": "1", "y": "2"}
    }
  `), pp)

  fmt.Println("err ", err )
  //fmt.Println("--------------------------------------")
  //fmt.Println("PARSE:", fmt.Sprintf("%#v", pp))
  fmt.Println("--------------------------------------")
  fmt.Println("Friends:", fmt.Sprintf("%#v", pp.Friends))
  fmt.Println("--------------------------------------")
  fmt.Println("Type2:", fmt.Sprintf("%#v", pp.Type2))
  fmt.Println("Types:", fmt.Sprintf("%#v", pp.Types))
  fmt.Println("Types2:", fmt.Sprintf("%#v", pp.Types2))
  fmt.Println("Data3:", fmt.Sprintf("%#v", pp.Data3))
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


}

// go test -v  github.com/mobilemindtec/go-utils/tests -run TestResultsJson
func  TestResultsJson(t *testing.T) {
  var results interface{} = &[]*Age{ &Age{Age: 50}, }
  px := &Result{ Data: results }
  b, err := json.Encode(px)
  fmt.Println("---------------------------")
  fmt.Println(string(b), ", ERR = ", err)  
  fmt.Println("---------------------------")

  var results2 interface{} = []*Age{ &Age{Age: 50}, }
  px = &Result{ Data: results2 }
  b, err = json.Encode(px)
  fmt.Println("---------------------------")
  fmt.Println(string(b), ", ERR = ", err)  
  fmt.Println("---------------------------")

}

