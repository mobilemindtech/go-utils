package db

import (
	"fmt"
)

type Page struct {
  Offset int64
  Limit int64
  Search string
  Order string
  Sort string  
  FilterColumns map[string]interface{} 
  AndFilterColumns map[string]interface{}
}

func (this *Page) AddFilterColumn(columnName string, value interface{}) *Page{
	
	if this.FilterColumns == nil {
		this.FilterColumns = make(map[string]interface{} )
	}	

	this.FilterColumns[columnName] = value

	return this
}

func (this *Page) AddDefaultFilter(columnName string) *Page{	
	return this.AddFilterColumn(fmt.Sprintf("%v__icontains", columnName), this.Search)
}

func (this *Page) AddAndFilter(columnName string, value interface{}) *Page{	
	if this.AndFilterColumns == nil {
		this.AndFilterColumns = make(map[string]interface{} )
	}	

	this.AndFilterColumns[columnName] = value

	return this
}

func (this *Page) MakeDefaultSort() {
	if this.Order == "asc" {
		this.Order = ""
	} else {
		this.Order = "-"
	}	
}