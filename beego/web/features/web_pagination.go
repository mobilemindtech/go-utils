package features

import (
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/support"
)

type WebPagination struct {
	base trait.WebBaseInterface
}

func (this *WebPagination) InitWebPagination(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebPagination) GetPage() *db.Page {
	page := new(db.Page)

	var defaultLimit int64 = 25

	if this.base.IsJson() {

		if this.base.GetCtx().Input.IsPost() {

			jsonData := json.Try(this.base.GetRawBody()).OrNil()

			if jsonData != nil && jsonData.HasKey("limit") {
				page.Limit = jsonData.OptInt64("limit", defaultLimit)
				page.Offset = jsonData.OptInt64("offset", 0)
				page.Search = jsonData.GetString("search")
				page.Sort = jsonData.GetString("sort")
				if len(page.Sort) == 0 {
					page.Sort = jsonData.GetString("order_column")
				}
				page.Order = jsonData.GetString("order")
				if len(page.Order) == 0 {
					page.Order = jsonData.GetString("order_sort")
				}
				return page
			}

		}
	}

	page.Limit = support.StrToInt64(this.base.GetQuery("limit"))
	if page.Limit <= 0 {
		page.Limit = defaultLimit
	}

	page.Sort = this.base.GetQuery("sort")
	if len(page.Sort) == 0 {
		page.Sort = this.base.GetQuery("order_column")
	}

	page.Order = this.base.GetQuery("order")
	if len(page.Order) == 0 {
		page.Order = this.base.GetQuery("order_sort")
	}

	page.Offset = support.StrToInt64(this.base.GetQuery("offset"))
	page.Search = this.base.GetQuery("search")

	return page
}
