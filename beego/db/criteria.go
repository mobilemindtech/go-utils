package db

import (
	"github.com/astaxie/beego/orm"
	"reflect"
	"strings"
	"errors"
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
	Or
	AndOr
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

	criateriasOr []*Criteria
	criateriasAndOr []*Criteria
	criateriasAnd []*Criteria

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
	HasError bool

	Debug bool
}

func NewCriteria(session *Session, entity interface{}, entities interface{}) *Criteria {
	return &Criteria{ criaterias: []*Criteria{}, criateriasOr: []*Criteria{}, criateriasAnd: []*Criteria{}, criateriasAndOr: []*Criteria{}, Session: session, Result: entity, Results: entities, Tenant: session.Tenant  }
}

func NewCondition() *Criteria{
	return &Criteria{ criaterias: []*Criteria{}  }	
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

func (this *Criteria) Or(criteria *Criteria) *Criteria {		
	this.criateriasOr = append(this.criateriasOr, criteria)
	return this
}

func (this *Criteria) And(criteria *Criteria) *Criteria {		
	this.criateriasAnd = append(this.criateriasAnd, criteria)
	return this
}

func (this *Criteria) AndOr(criteria *Criteria) *Criteria {		
	this.criateriasAndOr = append(this.criateriasAndOr, criteria)
	return this
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

func (this *Criteria) List() *Criteria {
	return this.execute(CriteriaList)
}

func (this *Criteria) One() *Criteria {
	return this.execute(CriteriaOne)
}

func (this *Criteria) Count() *Criteria {
	return this.execute(CriteriaCount)
}

func (this *Criteria) Get(id int64) *Criteria {
	this.Eq("Id", id)
	return this.execute(CriteriaOne)
}

func (this *Criteria) Query() orm.QuerySeter {
	
	if this.query == nil {

	  entity := this.Result

	  if model, ok := entity.(Model); ok {
	    this.query = this.Session.Db.QueryTable(model.TableName())			
		} else {
			this.setError(errors.New("entity does not implements of Model")	)
		}
	}	

	return this.query
}

func (this *Criteria) SetDebug(debug bool) *Criteria {
	this.Debug = debug
	return this.execute(CriteriaOne)
}

func (this *Criteria) buildPage() {
	
	if this.Page != nil {
    
		switch this.Page.Sort {
			case "asc":
    		this.OrderAsc(this.Page.Sort)
    	case "desc":
    		this.OrderDesc(this.Page.Sort)
		}

    if this.Page.FilterColumns != nil && len(this.Page.FilterColumns) > 0 {
        
      if len(this.Page.FilterColumns) == 1 {
        
        for k, v := range this.Page.FilterColumns {
          this.Eq(k, v)
        }

      } else {

        cond := NewCondition()
        for k, v := range this.Page.FilterColumns {
          cond.Eq(k, v)          
        }        
        this.AndOr(cond)

      }

    }

    if this.Page.AndFilterColumns != nil && len(this.Page.AndFilterColumns) > 0 {      
      for k, v := range this.Page.AndFilterColumns {
        this.Eq(k, v)
      }      
    }
	}

}

func (this *Criteria) build(query orm.QuerySeter) orm.QuerySeter {
	

	condition := orm.NewCondition()

	for _, criteria := range this.criaterias {

		pathName := this.getPathName(criteria)
		
		cond := orm.NewCondition()
		
		switch criteria.Expression {

			case Ne, IsNotNull, NotIn:
				//query = query.Exclude(pathName, criteria.Value)
				cond = cond.AndNot(pathName, criteria.Value)
			case Between:
				b := orm.NewCondition()
				b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
				b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
				cond = cond.AndCond(b)
			default:
				cond = cond.And(pathName, criteria.Value)
				//query = query.Filter(pathName, criteria.Value)

		}

		if this.Debug {
			fmt.Println("*********************************************************")
			fmt.Println("** set condition default %v ", pathName)
			fmt.Println("*********************************************************")
		}

		condition = condition.AndCond(cond)
		
	}


	for _, c := range this.criateriasOr {
		
		cond := orm.NewCondition()

		for _, criteria := range c.criaterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
				case Ne, IsNotNull, NotIn:
					cond = cond.OrNot(pathName, criteria.Value)
				case Between:
					b := orm.NewCondition()
					b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
					b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
					cond = cond.OrCond(b)
				default:
					cond = cond.Or(pathName, criteria.Value)
			}

			if this.Debug {
				fmt.Println("*********************************************************")
				fmt.Println("** set condition or %v ", pathName)
				fmt.Println("*********************************************************")		
			}
		}


		condition = condition.OrCond(cond)

	}

	for _, c := range this.criateriasAnd {
		
		cond := orm.NewCondition()

		for _, criteria := range c.criaterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
				case Ne, IsNotNull, NotIn:
					cond = cond.AndNot(pathName, criteria.Value)
				case Between:
					cond = cond.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
					cond = cond.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
				default:
					cond = cond.And(pathName, criteria.Value)
			}
			if this.Debug {
				fmt.Println("*********************************************************")
				fmt.Println("** set condition and %v ", pathName)
				fmt.Println("*********************************************************")
			}
		}


		condition = condition.AndCond(cond)

	}

	for _, c := range this.criateriasAndOr {
		
		cond := orm.NewCondition()

		for _, criteria := range c.criaterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
				case Ne, IsNotNull, NotIn:
					cond = cond.OrNot(pathName, criteria.Value)
				case Between:
					b := orm.NewCondition()
					b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
					b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
					cond = cond.OrCond(b)
				default:
					cond = cond.Or(pathName, criteria.Value)
			}

			if this.Debug {
				fmt.Println("*********************************************************")
				fmt.Println("** set condition and or %v ", pathName)
				fmt.Println("*********************************************************")		
			}

		}

		condition = condition.AndCond(cond)

	}	

	query = query.SetCond(condition)
	

	return query

}

func (this *Criteria) getPathName(criteria *Criteria) string {
		pathName := criteria.Path

		if strings.Contains(criteria.Path, "icontains") {
			return pathName
		}
		
		switch criteria.Expression {

			case Eq:

			case Ne:

			case Le:
				pathName = fmt.Sprintf("%v__lte", criteria.Path)
			case Lt:
				pathName = fmt.Sprintf("%v__lt", criteria.Path)
			case Ge:
				pathName = fmt.Sprintf("%v__gte", criteria.Path)
			case Gt:
				pathName = fmt.Sprintf("%v__gt", criteria.Path)
			case Like:

				switch criteria.Match {
					case Exact:
						pathName = fmt.Sprintf("%v__exact", criteria.Path)
					case IExact:
						pathName = fmt.Sprintf("%v__iexact", criteria.Path)
					case StartsWith:
						pathName = fmt.Sprintf("%v__startswith", criteria.Path)
					case IStartsWith:
						pathName = fmt.Sprintf("%v__istartswith", criteria.Path)
					case EndsWith:
						pathName = fmt.Sprintf("%v__endswith", criteria.Path)
					case IEndsWith:
						pathName = fmt.Sprintf("%v__iendswith", criteria.Path)
					case Anywhare:
						pathName = fmt.Sprintf("%v__contains", criteria.Path)
					case IAnywhare:
						pathName = fmt.Sprintf("%v__icontains", criteria.Path)
				}

			case IsNull:
				pathName = fmt.Sprintf("%v__isnull", criteria.Path)
			case IsNotNull:
				pathName = fmt.Sprintf("%v__isnull", criteria.Path)
			case Between:

			case In:
				pathName = fmt.Sprintf("%v__in", criteria.Path)
			case NotIn:
				pathName = fmt.Sprintf("%v__in", criteria.Path)

		}	

		return pathName
}

func (this *Criteria) execute(resultType CriteriaResult) *Criteria{

  query := this.Query()
  
  if this.Limit > 0 {
  	query = query.Limit(this.Limit).Offset(this.Offset)   
	}

  if this.Tenant != nil {
    query = query.Filter("Tenant", this.Tenant)
  }    

  this.buildPage()
  query = this.build(query)

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
  			this.setError(errors.New("Results can't be nil"))
  			return this
  		}

  		err := this.Session.ToList(query, this.Results)

  		this.setError(err)

  		this.Any = reflect.ValueOf(this.Results).Elem().Len() > 0

  	case CriteriaOne:

  		err := this.Session.ToOne(query, this.Result)

  		if err != orm.ErrNoRows {
	  		this.setError(err)
    	} else {
    		this.setError(nil)
    	}

			if model, ok := this.Result.(Model); ok {
				this.Any = model.IsPersisted()
			}

  	case CriteriaCount:  

  		count, err := this.Session.ToCount(query)

  		this.Count64 = count
  		this.Count32 = int(count) 

  		this.Any = count > 0

  		this.setError(err)
  }

  return this
    
}

func (this *Criteria) setError(err error) {
	if err != nil && this.Error == nil{
		this.HasError = true
		this.Error = err
	}

	this.query = nil
}