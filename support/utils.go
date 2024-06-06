package support

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/leekchan/accounting"
	"github.com/mobilemindtec/go-utils/v2/maps"
)

func FilterNumber(text string) string {
	re := regexp.MustCompile("[0-9]+")
	result := re.FindAllString(text, -1)
	number := ""
	for _, s := range result {
		number += s
	}

	return number
}

func IsEmpty(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}

func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func SliceCopyAndSortOfStrings(arr []string) []string {
	tmpArr := make([]string, len(arr))
	copy(tmpArr, arr)
	sort.Strings(tmpArr)
	return tmpArr
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// troca , por .(ponto), posi alterei o js maskMoney pra #.###,##
func NormalizeSemicolon(ctx *context.Context, keys ...string) {
	for _, key := range keys {
		if _, ok := ctx.Request.Form[key]; ok {
			ctx.Request.Form[key][0] = strings.Replace(ctx.Request.Form[key][0], ",", "", -1)
		}
	}
}

func RemoveAllSemicolon(ctx *context.Context, keys ...string) {
	for _, key := range keys {
		if _, ok := ctx.Request.Form[key]; ok {
			ctx.Request.Form[key][0] = strings.Replace(ctx.Request.Form[key][0], ",", "", -1)
		}
	}
}

func SetFormDefaults(ctx *context.Context, vals ...interface{}) {
	for k, v := range maps.Of[string, string](vals...) {
		SetFormDefault(k, v, ctx)
	}
}

func SetFormDefault(key string, defVal string, ctx *context.Context) {
	if _, ok := ctx.Request.Form[key]; ok {

		val := ctx.Request.Form[key][0]

		if len(strings.TrimSpace(val)) == 0 {
			ctx.Request.Form[key][0] = defVal
		}

	}
}

func FormatMoney(number float64) string {
	ac := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ",", Decimal: "."}
	return ac.FormatMoney(number)
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func NumberMask(text string, maskApply string) string {

	re := regexp.MustCompile("[0-9]+")
	results := re.FindAllString(text, -1)
	text = strings.Join(results[:], ",")

	var newText string
	var j int

	for i := 0; i < len(maskApply); i++ {

		m := maskApply[i]

		if j >= len(text) {
			newText += string(m)
			continue
		}

		c := text[j]

		if re.MatchString(string(c)) {
			if re.MatchString(string(m)) {
				newText += string(c)
				j++
			} else {
				newText += string(m)
			}
		}
	}

	return newText
}

func DateToTheEndOfDay(timeArg time.Time) time.Time {
	returnTime := timeArg.Local().Add(time.Hour*time.Duration(23) +
		time.Minute*time.Duration(59) +
		time.Second*time.Duration(59))
	return returnTime
}

func NumberMaskReverse(text string, maskApply string) string {

	re := regexp.MustCompile("[0-9]+")
	results := re.FindAllString(text, -1)
	text = strings.Join(results[:], ",")
	text = Reverse(text)

	var newText string
	var j int

	for i := len(maskApply) - 1; i >= 0; i-- {

		m := maskApply[i]

		if j >= len(text) {
			newText += string(m)
			continue
		}

		c := text[j]

		if re.MatchString(string(c)) {
			if re.MatchString(string(m)) {
				newText += string(c)
				j++
			} else {
				newText += string(m)
			}
		}
	}

	return Reverse(newText)
}

func StrToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func StrToInt64(s string) int64 {
	return int64(StrToInt(s))
}

func AnyToInt(s interface{}) int {

	if i, ok := s.(int); ok {
		return i
	}

	return StrToInt(s.(string))
}

func AnyToInt64(s interface{}) int64 {
	return int64(AnyToInt(s))
}

func StrToFloat(s string) float32 {
	i, _ := strconv.ParseFloat(s, 32)
	return float32(i)
}

func StrToFloat64(s string) float64 {
	i, _ := strconv.ParseFloat(s, 32)
	return i
}

func AnyToFloat(s interface{}) float32 {

	if i, ok := s.(float32); ok {
		return i
	}

	return StrToFloat(s.(string))
}

func AnyToFloat64(s interface{}) float64 {
	return float64(AnyToInt(s))
}