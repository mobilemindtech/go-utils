package db

import (
	"github.com/astaxie/beego/orm"
	"errors"
	"reflect"
	"fmt"
)

type CriteriaExpression int
type CriteriaLikeMatch int
type CriteriaResult int

const (

	Eq CriteriaExpression = 1 + iota
	Ne
	Le
	Lt
	Ge
	Gt
	Like
	Between
	IsNull
	IsNotNull
	In
	NotIn
)

const (
	Exact CriteriaLikeMatch = 1 + iota
	IExact
	StartsWith
	IStartsWith
	EndsWith
	IEndsWith
	Anywhare
	IAnywhare
)

const (
	CriteriaList CriteriaResult = 1 + iota
	CriteriaOne
	CriteriaCount
)

type CriteriaOrder struct {
	Path string
	Desc bool
}

type Criteria struct {

	Path string
	Value interface{}
	Value2 interface{}
	Expression CriteriaExpression

	Match CriteriaLikeMatch

	InValues []interface{}

	criaterias []*Criteria
	orderBy []*CriteriaOrder

	Result interface{}
	Results interface{}

	Page *Page

	Error error
	
	Count32 int
	Count64 int64

	Limit int64
	Offset int64

	Session *Session

	query orm.QuerySeter

	Tenant interface{}

	Any bool
}

func NewCriteria(session *Session, entity interface{}, entities interface{}) *Criteria {
	return &Criteria{ criaterias: []*Criteria{}, Session: session, Result: entity, Results: entities, Tenant: session.Tenant  }
}

func (this *Criteria) add(path string, value interface{}, expression CriteriaExpression) *Criteria{
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Value: value, Expression: expression } )
	return this
}	

func (this *Criteria) SetEntity(entity interface{}) *Criteria {
	this.Result = entity
	return this
}

func (this *Criteria) SetResult(result interface{}) *Criteria {
	this.Result = result
	return this
}

func (this *Criteria) SetResults(results []interface{}) *Criteria {
	this.Results = results
	return this
}

func (this *Criteria) SetPage(page *Page) *Criteria {
	this.Page = page

	this.Limit = page.Limit
	this.Offset = page.Offset

	return this
}

func (this *Criteria) SetTenant(tenant interface{}) *Criteria {
	this.Tenant = tenant
	return this
}


func (this *Criteria) Eq(path string, value interface{}) *Criteria {		
	return this.add(path, value, Eq)
}

func (this *Criteria) Ne(path string, value interface{}) *Criteria {		
	return this.add(path, value, Ne)	
}

func (this *Criteria) Le(path string, value interface{}) *Criteria {		
	return this.add(path, value, Le)	
}

func (this *Criteria) Lt(path string, value interface{}) *Criteria {		
	return this.add(path, value, Lt)	
}

func (this *Criteria) Ge(path string, value interface{}) *Criteria {		
	return this.add(path, value, Ge)	
}

func (this *Criteria) Gt(path string, value interface{}) *Criteria {		
	return this.add(path, value, Gt)	
}

func (this *Criteria) Like(path string, value interface{}) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Value: value, Expression: Like, Match: IAnywhare } )
	return this
}

func (this *Criteria) LikeMatch(path string, value interface{}, likeMatch CriteriaLikeMatch) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Value: value, Expression: Like, Match: likeMatch } )
	return this
}


func (this *Criteria) Between(path string, value interface{}, value2 interface{}) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Value: value, Value2: value2, Expression: Between } )
	return this
}


func (this *Criteria) IsNull(path string) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Expression: IsNull } )
	return this
}

func (this *Criteria) IsNotNull(path string) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Expression: IsNotNull } )
	return this
}

func (this *Criteria) In(path string, values ...interface{}) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Expression: In, InValues: values } )
	return this
}

func (this *Criteria) NotIn(path string, values ...interface{}) *Criteria {		
	this.criaterias = append(this.criaterias, &Criteria{ Path: path, Expression: NotIn, InValues: values } )
	return this
}

func (this *Criteria) OrderAsc(path string) *Criteria{
	this.orderBy = append(this.orderBy, &CriteriaOrder{ Path: path })
	return this
}

func (this *Criteria) OrderDesc(path string) *Criteria{
	this.orderBy = append(this.orderBy, &CriteriaOrder{ Path: path, Desc: true })
	return this
}

func (this *Criteria) Build(query orm.QuerySeter) orm.QuerySeter {
	

	for _, criteria := range this.criaterias {

		switch criteria.Expression {

			case Eq:
				query = query.Filter(criteria.Path, criteria.Value)
			case Ne:
				query = query.Exclude(criteria.Path, criteria.Value)
			case Le:
				query = query.Filter(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value)
			case Lt:
				query = query.Filter(fmt.Sprintf("%v__lt", criteria.Path), criteria.Value)
			case Ge:
				query = query.Filter(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
			case Gt:
				query = query.Filter(fmt.Sprintf("%v__gt", criteria.Path), criteria.Value)
			case Like:

				switch criteria.Match {
					case Exact:
						query = query.Filter(fmt.Sprintf("%v__exact", criteria.Path), criteria.Value)
					case IExact:
						query = query.Filter(fmt.Sprintf("%v__iexact", criteria.Path), criteria.Value)
					case StartsWith:
						query = query.Filter(fmt.Sprintf("%v__startswith", criteria.Path), criteria.Value)
					case IStartsWith:
						query = query.Filter(fmt.Sprintf("%v__istartswith", criteria.Path), criteria.Value)
					case EndsWith:
						query = query.Filter(fmt.Sprintf("%v__endswith", criteria.Path), criteria.Value)
					case IEndsWith:
						query = query.Filter(fmt.Sprintf("%v__iendswith", criteria.Path), criteria.Value)
					case Anywhare:
						query = query.Filter(fmt.Sprintf("%v__contains", criteria.Path), criteria.Value)
					case IAnywhare:
						query = query.Filter(fmt.Sprintf("%v__icontains", criteria.Path), criteria.Value)
				}

			case IsNull:
				query = query.Filter(fmt.Sprintf("%v__isnull", criteria.Path), criteria.Value)			
			case IsNotNull:
				query = query.Exclude(fmt.Sprintf("%v__isnull", criteria.Path), criteria.Value)
			case Between:
				query = query.Filter(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
				query = query.Filter(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
			case In:
				query = query.Filter(fmt.Sprintf("%v__in", criteria.Path), criteria.InValues)
			case NotIn:
				query = query.Filter(fmt.Sprintf("%v__in", criteria.Path), criteria.InValues)

		}
		
	}

	return query

}

func (this *Criteria) List() *Criteria {
	return this.execute(CriteriaList)
}

func (this *Criteria) One() *Criteria {
	return this.execute(CriteriaOne)
}

func (this *Criteria) Count() *Criteria {
	return this.execute(CriteriaCount)
}

func (this *Criteria) Query() orm.QuerySeter {
	
	if this.query == nil {

	  entity := this.Result

	  if model, ok := entity.(Model); ok {

	    this.query = this.Session.Db.QueryTable(model.TableName())			

		} else {

			this.Error = errors.New("entity does not implements of Model")	

		}
	}	

	return this.query
}

func (this *Criteria) execute(resultType CriteriaResult) *Criteria{

  query := this.Query()
  
  if this.Limit > 0 {
  	query = query.Limit(this.Limit).Offset(this.Offset)   
	}

	if this.Page != nil {
    if this.Page.Sort != "" { 
      query = query.OrderBy(fmt.Sprintf("%v%v", this.Page.Order, this.Page.Sort))
    }

    if this.Page.FilterColumns != nil && len(this.Page.FilterColumns) > 0 {
        
      if len(this.Page.FilterColumns) == 1 {
        for k, v := range this.Page.FilterColumns {
          query = query.Filter(k, v)
        }
      } else {
        cond := orm.NewCondition()
        for k, v := range this.Page.FilterColumns {
          cond = cond.Or(k, v)          
        }        
        query = query.SetCond(orm.NewCondition().AndCond(cond))
      }

    }

    if this.Page.AndFilterColumns != nil && len(this.Page.AndFilterColumns) > 0 {      
      for k, v := range this.Page.AndFilterColumns {
        query = query.Filter(k, v)
      }      
    }
	}

  if this.Tenant != nil {
    query.Filter("Tenant", this.Tenant)
  }    

  query = this.Build(query)

  switch resultType {    	

  	case CriteriaList:

  		for _, order := range this.orderBy {
  			if order.Desc {
  				query = query.OrderBy(fmt.Sprintf("-%v", order.Path))    				
  			} else {
  				query = query.OrderBy(fmt.Sprintf(order.Path))
  			}
  		}

  		if this.Results == nil {
  			this.Error = errors.New("Results can't be nil")
  			return this
  		}

  		this.Error = this.Session.ToList(query, this.Results)

  		s := reflect.ValueOf(this.Results).Elem()

  		this.Any = s.Len() > 0

  	case CriteriaOne:
  		this.Error = this.Session.ToOne(query, this.Result)

  		if this.Error == orm.ErrNoRows {
      	this.Error = nil
      	this.Result = nil
    	} else {
    		this.Any = true
    	}

  	case CriteriaCount:    		
  		this.Count64, this.Error = this.Session.ToCount(query)
  		this.Count32 = int(this.Count64) 

  		this.Any = this.Count32 > 0
  }

  this.query = nil

  return this
    
}