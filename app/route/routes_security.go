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
    for route, value := range routes{
    	fmt.Println(" path: %v, role: %v", route, value)
  	}
  	fmt.Println("****************** routes ******************")
}

func IsRouteAuthorized(ctx *context.Context, currentAuthUserRoles []string) bool {


	var routeConfigured bool
	requestedUrl := ctx.Input.URL()


	fmt.Println("** check route %v, roles %v", requestedUrl, currentAuthUserRoles)


	for route, value := range routes{

		var allow bool
		
		// verifica se é curinga
		if strings.HasSuffix(route, "/*") {

			
			var hasUniqueRule bool 

			// verifica se tem uma regra unida para esse path
			for _, it := range routes {
				if requestedUrl == it {
					hasUniqueRule = true
				}
			}

			if !hasUniqueRule {
				// como tem o curinga * verifica se a base path é a mesma

				allow = checkCuringa(route, requestedUrl)
				
			}

		}

		if requestedUrl == route  || allow {

			routeConfigured = true
			roleNames := value.(string)

			if roleNames == "anonymous" {
				fmt.Println("** route %v anonymous allowed", requestedUrl)
				return true
			}

			if roleNames == "authenticated" {
				// authService is nil, so not auth
				fmt.Println("** route %v authenticated user allowed", requestedUrl)
				return len(currentAuthUserRoles) > 0
			}

			values := strings.Split(roleNames, ",")

			for _, roleName := range values {
				for _, role := range currentAuthUserRoles {
					if role == roleName {
						fmt.Println("** route %v authenticated user role allowed", requestedUrl)
						return true
					} else {
            fmt.Println("** route %v not authenticated user role allowed for role ", requestedUrl, roleName)
          }
				}
			}

		}

	}

	if !routeConfigured {
		fmt.Println("** route %v not is configured in routes.json", requestedUrl)
	}

	fmt.Println("** route %v not allowed", requestedUrl)
	return false
}

func checkCuringa(route string, url string) bool {
	
	routesSplited := strings.Split(route, "/")
	requestedUrlSplited := strings.Split(url, "/")

	routesSplited = routesSplited[:len(routesSplited)-1]

	allow := false

	if len(routesSplited) - 1 > 0 && len(requestedUrlSplited) > 0 {

		allow = true


		for index, it := range routesSplited {

			if index > len(requestedUrlSplited) -1 {
				allow = false
				break
			}
			
			if it != requestedUrlSplited[index] {
				allow = false
			}

		}

	}	

	return allow

}