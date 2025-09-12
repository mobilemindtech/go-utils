package web

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"
	"github.com/leekchan/accounting"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/filters"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/satori/go.uuid"
	"strconv"
	"strings"
	"time"
)

var (
	langTypes []string // Languages that are supported.
	//datetimeLayout = "02/01/2006 15:04:05"
	//timeLayout = "10:25"
	//dateLayout = "02/01/2006"
	//jsonDateLayout = "2006-01-02T15:04:05-07:00"
)

type RecoverInfo struct {
	Error      string
	StackTrace string
}

type NestPreparer interface {
	NestPrepare()
}

type NestFinisher interface {
	NestFinish()
}

type NestRecover interface {
	NextOnRecover(info *RecoverInfo)
}

type NestWebController interface {
	WebControllerLoadModels()
	WebControllerCreateSession() *db.Session
}

func LoadFuncs() {
	inc := func(i int) int {
		return i + 1
	}

	hasError := func(args map[string]string, key string) string {
		if args[key] != "" {
			return "has-error"
		}
		return ""
	}

	errorMsg := func(args map[string]string, key string) string {
		return args[key]
	}

	currentYaer := func() string {
		return strconv.Itoa(time.Now().Year())
	}

	formatMoney := func(number float64) string {
		ac := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ",", Decimal: "."}
		return ac.FormatMoney(number)
	}

	isZeroDate := func(date time.Time) bool {
		return time.Time.IsZero(date) || date.Year() <= 1900
	}

	formatDate := func(date time.Time) string {

		if !isZeroDate(date) {
			return date.Format("02/01/2006")
		}
		return "-"
	}

	formatTime := func(date time.Time) string {
		if !isZeroDate(date) {
			return date.Format("15:04")
		}
		return "-"
	}

	formatDateTime := func(date time.Time) string {
		if !isZeroDate(date) {
			return date.Format("02/01/2006 15:04")
		}
		return "-"
	}

	dateFormat := func(date time.Time, layout string) string {
		if !isZeroDate(date) {
			return date.Format(layout)
		}
		return "-"
	}

	getNow := func(layout string) string {
		return time.Now().Format(layout)
	}

	getYear := func() string {
		return time.Now().Format("2006")
	}

	formatBoolean := func(b bool, wrapLabel bool) string {
		var s string
		if b {
			s = "Sim"
		} else {
			s = "Não"
		}
		if wrapLabel {
			var class string
			if b {
				class = "info"
			} else {
				class = "danger"
			}
			val := "<span class='label label-" + class + "'>" + s + "</span>"
			s = val
		}
		return s
	}

	formatDecimal := func(number float64) string {
		ac := accounting.Accounting{Symbol: "", Precision: 2, Thousand: ",", Decimal: "."}
		return ac.FormatMoney(number)
	}

	sum := func(numbers ...float64) float64 {
		total := 0.0
		for i, it := range numbers {
			if i == 0 {
				total = it
			} else {
				total += it
			}
		}
		return total
	}

	subtract := func(numbers ...float64) float64 {
		total := 0.0
		for i, it := range numbers {
			if i == 0 {
				total = it
			} else {
				total -= it
			}
		}
		return total
	}

	mult := func(numbers ...float64) float64 {
		total := 0.0
		for i, it := range numbers {
			if i == 0 {
				total = it
			} else {
				total *= it
			}
		}
		return total
	}

	numberMask := func(text interface{}, mask string) string {
		return support.NumberMask(fmt.Sprintf("%v", text), mask)
	}

	numberMaskReverse := func(text interface{}, mask string) string {
		return support.NumberMaskReverse(fmt.Sprintf("%v", text), mask)
	}

	uuidGen := func() string {
		return uuid.NewV4().String()
	}

	timeDiff := func(start time.Time, end time.Time, label string) string {
		if isZeroDate(start) || isZeroDate(end) {
			return "-"
		}
		diff := end.Sub(start)
		return fmt.Sprintf("%v %v", support.ToFixed(diff.Hours(), 2), label)
	}

	beego.AddFuncMap("is_zero_date", isZeroDate)
	beego.AddFuncMap("uuid", uuidGen)
	beego.AddFuncMap("inc", inc)
	beego.AddFuncMap("has_error", hasError)
	beego.AddFuncMap("error_msg", errorMsg)
	beego.AddFuncMap("current_yaer", currentYaer)
	beego.InsertFilter("*", beego.BeforeRouter, filters.FilterMethod) // enable put
	beego.AddFuncMap("format_boolean", formatBoolean)
	beego.AddFuncMap("format_date", formatDate)
	beego.AddFuncMap("format_time", formatTime)
	beego.AddFuncMap("date_format", dateFormat)
	beego.AddFuncMap("time_diff", timeDiff)
	beego.AddFuncMap("get_now", getNow)
	beego.AddFuncMap("get_year", getYear)
	beego.AddFuncMap("format_date_time", formatDateTime)
	beego.AddFuncMap("format_money", formatMoney)
	beego.AddFuncMap("format_decimal", formatDecimal)
	beego.AddFuncMap("sum", sum)
	beego.AddFuncMap("subtract", subtract)
	beego.AddFuncMap("mult", mult)

	beego.AddFuncMap("mask", numberMask)
	beego.AddFuncMap("mask_reverse", numberMaskReverse)
}

func LoadIl8n() {
	beego.AddFuncMap("i18n", i18n.Tr)
	logs.SetLevel(logs.LevelDebug)

	// Initialize language type list.
	types, _ := beego.AppConfig.String("lang_types")
	langTypes = strings.Split(types, "|")

	logs.Info(" langTypes %v", langTypes)

	// Load locale files according to language types.
	for _, lang := range langTypes {
		if err := i18n.SetMessage(lang, "conf/i18n/"+"locale_"+lang+".ini"); err != nil {
			logs.Error("Fail to set message file:", err)
			return
		}
	}
}
