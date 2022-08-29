package route

import (
	"github.com/beego/beego/v2/server/web/context"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/core/logs"
	"encoding/json"
	"io/ioutil"
	"strings"
	"fmt"
)


var routes map[string]interface{}

func init() {
    
    configpath := fmt.Sprintf("%v/conf/routes.json", beego.WorkPath)
    file, err := ioutil.ReadFile(configpath)
    if err != nil {
        logs.Debug("error open route config file %v: %v\n", configpath, err)
        return
    }

    data := make(map[string]interface{})
    
    err = json.Unmarshal(file, &data)
    if err != nil {
        logs.Debug("JSON error parse route config file: %v\n", err)
        return
    }

    routes = data["routes"].(map[string]interface{})	

    logs.Debug("****************** routes ******************")
    for route, value := range routes{
    	logs.Debug(" path: %v, role: %v", route, value)
  	}
  	logs.Debug("****************** routes ******************")
}

func IsRouteAuthorized(ctx *context.Context, currentAuthUserRoles []string) bool {


	var routeConfigured bool
	requestedUrl := ctx.Input.URL()

	logs.Debug(fmt.Sprintf("check route %v, roles %v", requestedUrl, currentAuthUserRoles))


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
				logs.Debug(fmt.Sprintf("route %v anonymous allowed", requestedUrl))
				return true
			}

			if roleNames == "authenticated" {
				// authService is nil, so not auth
				logs.Debug(fmt.Sprintf("route %v authenticated user allowed", requestedUrl))
				return len(currentAuthUserRoles) > 0
			}

			values := strings.Split(roleNames, ",")

			for _, roleName := range values {
				for _, role := range currentAuthUserRoles {
					if role == roleName {
						logs.Debug(fmt.Sprintf("route %v authenticated user role allowed", requestedUrl))
						return true
					} else {
            logs.Debug(fmt.Sprintf("route %v not authenticated user role allowed for role ", requestedUrl, roleName))
          }
				}
			}

		}

	}

	if !routeConfigured {
		logs.Debug(fmt.Sprintf("route %v not is configured in routes.json", requestedUrl))
	}

	logs.Debug(fmt.Sprintf("route %v not allowed", requestedUrl))
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