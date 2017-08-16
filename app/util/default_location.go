
package util

import (
	"time"
)

func GetDefaultLocation() *time.Location {
	location, _ := time.LoadLocation("America/Sao_Paulo") 
	return location
}