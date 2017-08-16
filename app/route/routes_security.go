package route

import (
	"github.com/astaxie/beego/context"
  "encoding/json"
  "io/ioutil"
  "strings"
  "fmt"
)


var routes map[string]interface{}

func init() {

    file, err := ioutil.ReadFile("./conf/routes.json")
    if err != nil {
        fmt.Printf("error on open file ./conf/routes.json: %v\n", err)
        return
    }

    data := make(map[string]interface{})
    
    err = json.Unmarshal(file, &data)
    if err != nil {
        fmt.Printf("JSON error: %v\n", err)
        return
    }

    routes = data["routes"].(map[string]interface{})	

    fmt.Println("****************** routes ******************")
    for key, value := range routes{
    	fmt.Println(" path: %v, role: %v", key, value)
  	}
  	fmt.Println("****************** routes ******************")
}

func IsRouteAuthorized(ctx *context.Context, currentAuthUserRoles []string) bool {


	var routeConfigured bool
	route := ctx.Input.URL()



	for key, value := range routes{

		var allow bool
		
		// verifica se é curinga
		if strings.HasSuffix(key, "/*") {

			base := strings.Split(key, "/")[1]
			var hasUniqueRule bool 

			// verifica se tem uma regra unida para esse path
			for _, it := range routes {
				if route == it {
					hasUniqueRule = true
				}
			}

			if !hasUniqueRule {
				// como tem o curinga * verifica se a base path é a mesma
				allow = strings.Split(route, "/")[1] == base
			}

		}

		if route == key  || allow {

			routeConfigured = true
			roleNames := value.(string)

			if roleNames == "anonymous" {
				return true
			}

			if roleNames == "authenticated" {
				// authService is nil, so not auth
				return len(currentAuthUserRoles) > 0
			}

			values := strings.Split(roleNames, ",")

			for _, roleName := range values {
				for _, role := range currentAuthUserRoles {
					if role == roleName {
						return true
					}
				}
			}

		}

	}

	if !routeConfigured {
		fmt.Println("** route %v not is configured in routes.json", route)
	}

	return false
}