package support 

import (
	"regexp"
	"strings"
)


func FilterNumber(text string) string{
	re := regexp.MustCompile("[0-9]+")
	result := re.FindAllString(text, -1)
	number := ""
	for _, s := range result {
		number += s
	}

	return number
}

func IsEmpty(text string) bool{
  return len(strings.TrimSpace(text)) == 0
}