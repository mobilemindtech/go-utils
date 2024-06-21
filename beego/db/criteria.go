package db

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/v2/optional"
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
	NotLike
	Between
	IsNull
	IsNotNull
	In
	NotIn
	Or
	AndOr
	OrAnd
	AndOrAnd
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
	CriteriaListAndCount
	CriteriaOne
	CriteriaCount
	CriteriaUpdate
	CriteriaDelete
	CriteriaExists
	CriteriaAggregateOne
	CriteriaAggregateList
)

type CriteriaOrder struct {
	Path string
	Desc bool
}

type CriteriaSet struct {
	Criterias []*Criteria
}

func NewCriteriaSet() *CriteriaSet {
	return &CriteriaSet{Criterias: []*Criteria{}}
}

func NewCriteriaSetWithConditions(criterias ...*Criteria) *CriteriaSet {
	item := &CriteriaSet{Criterias: []*Criteria{}}
	for _, it := range criterias {
		item.Criterias = append(item.Criterias, it)
	}
	return item
}

func (this *CriteriaSet) AddCriteria(criterias ...*Criteria) *CriteriaSet {
	for _, it := range criterias {
		this.Criterias = append(this.Criterias, it)
	}
	return this
}

type Criteria struct {
	Path       string
	Value      interface{}
	Value2     interface{}
	Expression CriteriaExpression

	Match CriteriaLikeMatch

	InValues []interface{}

	criterias []*Criteria
	orderBy   []*CriteriaOrder

	criteriasOr       []*Criteria
	criteriasAndOr    []*Criteria
	criteriasAndOrAnd []*CriteriaSet
	criteriasOrAnd    []*Criteria
	criteriasAnd      []*Criteria

	Result  interface{}
	Results interface{}

	criteriaType CriteriaResult

	UpdateParams map[string]interface{}

	Page *Page
	searchPaths []string
	searchValue string

	Error error

	Count32 int
	Count64 int64

	Limit  int64
	Offset int64

	Session *Session

	query orm.QuerySeter

	RelatedSelList []string

	Any      bool
	Empty    bool
	HasError bool

	ForceAnd bool
	ForceOr  bool
	Distinct bool

	Debug bool

	tenantCopy interface{}

	aggregate string
	groupBy   string

	resultAggregate  interface{}
	resultsAggregate interface{}
}

func NewCriteria(session *Session, entity interface{}, entities interface{}) *Criteria {
	return &Criteria{
		criterias: []*Criteria{},
		criteriasOr: []*Criteria{},
		criteriasAnd: []*Criteria{},
		criteriasAndOr: []*Criteria{},
		criteriasAndOrAnd: []*CriteriaSet{},
		criteriasOrAnd: []*Criteria{},
		Session: session,
		Result: entity,
		Results: entities,
		RelatedSelList: []string{},
		searchPaths: []string{},
	}
}

func NewCondition() *Criteria {
	return &Criteria{criterias: []*Criteria{}}
}

func (this *Criteria) CopyConditions(c *Criteria) {

	for _, it := range c.criterias {
		this.criterias = append(c.criterias, it)
	}

	for _, it := range c.criteriasOr {
		this.criteriasOr = append(c.criteriasOr, it)
	}

	for _, it := range c.criteriasAndOr {
		this.criteriasAndOr = append(c.criteriasAndOr, it)
	}

	for _, it := range c.criteriasAndOrAnd {
		this.criteriasAndOrAnd = append(c.criteriasAndOrAnd, it)
	}

	for _, it := range c.criteriasOrAnd {
		this.criteriasOrAnd = append(c.criteriasOrAnd, it)
	}

	for _, it := range c.criteriasAnd {
		this.criteriasAnd = append(c.criteriasAnd, it)
	}

}

func (this *Criteria) IsOne() bool {
	return this.criteriaType == CriteriaOne
}

func (this *Criteria) IsList() bool {
	return this.criteriaType == CriteriaList
}

func (this *Criteria) IsListAndCount() bool {
	return this.criteriaType == CriteriaListAndCount
}

func (this *Criteria) IsCount() bool {
	return this.criteriaType == CriteriaCount
}

func (this *Criteria) IsExists() bool {
	return this.criteriaType == CriteriaExists
}

func (this *Criteria) SetDefaults() *Criteria {
	this.RelatedSelList = []string{}
	return this.clearConditions()
}

func (this *Criteria) clearConditions() *Criteria {
	this.criterias = []*Criteria{}
	this.criteriasOr = []*Criteria{}
	this.criteriasAnd = []*Criteria{}
	this.criteriasAndOr = []*Criteria{}
	this.criteriasAndOrAnd = []*CriteriaSet{}
	this.criteriasOrAnd = []*Criteria{}
	return this
}

func (this *Criteria) add(path string, value interface{}, expression CriteriaExpression, forceAnd bool, forceOr bool) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Value: value, Expression: expression, ForceAnd: forceAnd, ForceOr: forceOr})
	return this
}

func (this *Criteria) SetEntity(entity interface{}) *Criteria {
	this.Result = entity
	return this
}

func (this *Criteria) RunWithTenant(tenant interface{}, runner func(c *Criteria)) {
	this.Session.RunWithTenant(tenant, func() {
		runner(this)
	})
}

func (this *Criteria) WithTenant(tenant interface{}) *Criteria {
	this.tenantCopy = tenant
	this.Session.SetTenant(tenant)
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

func (this *Criteria) SetOffset(offset int64) *Criteria {
	this.Offset = offset
	return this
}

func (this *Criteria) SetLimit(limit int64) *Criteria {
	this.Limit = limit
	return this
}

func (this *Criteria) SetRelatedSel(related string) *Criteria {
	this.RelatedSelList = append(this.RelatedSelList, related)
	return this
}

func (this *Criteria) SetRelatedsSel(relateds ...string) *Criteria {
	for _, it := range relateds {
		this.RelatedSelList = append(this.RelatedSelList, it)
	}
	return this
}

func (this *Criteria) SetDistinct() *Criteria {
	this.Distinct = true
	return this
}

func (this *Criteria) SearchVal(value string) *Criteria {
	this.searchValue = value
	return this
}

func (this *Criteria) SearchCols(paths ...string) *Criteria {
	for _, path := range paths {
		this.searchPaths = append(this.searchPaths, path)
	}
	return this
}

func (this *Criteria) Eq(path string, value interface{}) *Criteria {
	return this.add(path, value, Eq, false, false)
}

func (this *Criteria) EqAnd(path string, value interface{}) *Criteria {
	return this.add(path, value, Eq, true, false)
}

func (this *Criteria) Ne(path string, value interface{}) *Criteria {
	return this.add(path, value, Ne, false, false)
}

func (this *Criteria) Le(path string, value interface{}) *Criteria {
	return this.add(path, value, Le, false, false)
}

func (this *Criteria) Lt(path string, value interface{}) *Criteria {
	return this.add(path, value, Lt, false, false)
}

func (this *Criteria) Ge(path string, value interface{}) *Criteria {
	return this.add(path, value, Ge, false, false)
}

func (this *Criteria) Gt(path string, value interface{}) *Criteria {
	return this.add(path, value, Gt, false, false)
}

func (this *Criteria) Or(criteria *Criteria) *Criteria {
	this.criteriasOr = append(this.criteriasOr, criteria)
	return this
}

func (this *Criteria) And(criteria *Criteria) *Criteria {
	this.criteriasAnd = append(this.criteriasAnd, criteria)
	return this
}

func (this *Criteria) AndOr(criteria *Criteria) *Criteria {
	this.criteriasAndOr = append(this.criteriasAndOr, criteria)
	return this
}

func (this *Criteria) AndOrAnd(criteriaSet *CriteriaSet) *Criteria {
	this.criteriasAndOrAnd = append(this.criteriasAndOrAnd, criteriaSet)
	return this
}

func (this *Criteria) OrAnd(criteria *Criteria) *Criteria {
	this.criteriasOrAnd = append(this.criteriasOrAnd, criteria)
	return this
}

func (this *Criteria) Like(path string, value interface{}) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Value: value, Expression: Like, Match: IAnywhare})
	return this
}

func (this *Criteria) NotLike(path string, value interface{}) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Value: value, Expression: NotLike, Match: IAnywhare})
	return this
}

func (this *Criteria) LikeMatch(path string, value interface{}, likeMatch CriteriaLikeMatch) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Value: value, Expression: Like, Match: likeMatch})
	return this
}

func (this *Criteria) NotLikeMatch(path string, value interface{}, likeMatch CriteriaLikeMatch) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Value: value, Expression: NotLike, Match: likeMatch})
	return this
}

func (this *Criteria) Between(path string, value interface{}, value2 interface{}) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Value: value, Value2: value2, Expression: Between})
	return this
}

func (this *Criteria) IsNull(path string) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Expression: IsNull})
	return this
}

func (this *Criteria) IsNotNull(path string) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Expression: IsNotNull})
	return this
}

func (this *Criteria) In(path string, values ...interface{}) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Expression: In, InValues: values})
	return this
}

func (this *Criteria) NotIn(path string, values ...interface{}) *Criteria {
	this.criterias = append(this.criterias, &Criteria{Path: path, Expression: NotIn, InValues: values})
	return this
}

func (this *Criteria) OrderAsc(paths ...string) *Criteria {
	for _, path := range paths {
		this.orderBy = append(this.orderBy, &CriteriaOrder{Path: path})
	}
	return this
}

func (this *Criteria) OrderDesc(paths ...string) *Criteria {
	for _, path := range paths {
		this.orderBy = append(this.orderBy, &CriteriaOrder{Path: path, Desc: true})
	}
	return this
}

func (this *Criteria) List() *Criteria {
	return this.execute(CriteriaList)
}

func (this *Criteria) GroupBy(s string) *Criteria {
	this.groupBy = s
	return this
}

func (this *Criteria) AggregateOne(s string, r interface{}) *Criteria {
	this.aggregate = s
	this.resultAggregate = r
	return this.execute(CriteriaAggregateOne)
}

func (this *Criteria) AggregateList(s string, r interface{}) *Criteria {
	this.aggregate = s
	this.resultsAggregate = r
	return this.execute(CriteriaAggregateList)
}

func (this *Criteria) ListAndCount() *Criteria {
	this.execute(CriteriaList)
	this.execute(CriteriaCount)
	this.criteriaType = CriteriaListAndCount
	return this
}

func (this *Criteria) One() *Criteria {
	return this.execute(CriteriaOne)
}

func (this *Criteria) Exists() bool {
	this.execute(CriteriaExists)
	return this.Any
}

func (this *Criteria) Count() *Criteria {
	return this.execute(CriteriaCount)
}

func (this *Criteria) Get(id int64) *Criteria {
	this.Eq("Id", id)
	return this.execute(CriteriaOne)
}

func (this *Criteria) Delete() *Criteria {
	return this.execute(CriteriaDelete)
}

func (this *Criteria) Update(args map[string]interface{}) *Criteria {
	this.UpdateParams = args
	return this.execute(CriteriaUpdate)
}

func (this *Criteria) Query() orm.QuerySeter {

	if this.query == nil {

		//entity := this.Result

		if model, ok := this.Result.(Model); ok {
			this.query = this.Session.GetDb().QueryTable(model.TableName())
		} else {
			this.SetError(errors.New("entity does not implements of Model"))
		}
	}

	return this.query
}

func (this *Criteria) SetDebug(debug bool) *Criteria {
	this.Debug = debug
	return this
}

func (this *Criteria) buildSearchPaths(paths []string, val string) {

	if len(paths) > 0 && len(val) > 0 {
		if len(paths) == 1 {
			for _, path := range paths {
				this.LikeMatch(path, val, IAnywhare)
			}
		} else {
			cond := NewCondition()
			for _, path := range paths {
				cond.LikeMatch(path, val, IAnywhare)
			}
			this.AndOr(cond)
		}

	}
}

func (this *Criteria) buildPage() {

	if this.Page != nil {

		if len(strings.TrimSpace(this.Page.Sort)) > 0 {
			switch this.Page.Order {
			case "asc":
				this.OrderAsc(this.Page.Sort)
			case "desc":
				this.OrderDesc(this.Page.Sort)
			}
		}

		if this.Page.FilterColumns != nil && len(this.Page.FilterColumns) > 0{

			if len(this.Page.FilterColumns) == 1 {

				for k, v := range this.Page.FilterColumns {
					this.Eq(k, v)
				}

			} else {

				cond := NewCondition()
				for k, v := range this.Page.FilterColumns {
					cond.Eq(k, v)
				}

				for k, v := range this.Page.TenantColumnFilter {
					cond.EqAnd(k, v)
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

func (this *Criteria) buildCriterias(criterias []*Criteria) *orm.Condition {
	condition := orm.NewCondition()

	for _, criteria := range criterias {

		pathName := this.getPathName(criteria)

		cond := orm.NewCondition()

		switch criteria.Expression {

		case In:
			cond = cond.And(pathName, criteria.InValues)
		case Ne, NotLike:
			cond = cond.AndNot(pathName, criteria.Value)
		case NotIn:
			cond = cond.AndNot(pathName, criteria.InValues)
		case IsNull:
			cond = cond.And(pathName, true)
		case IsNotNull:
			cond = cond.And(pathName, false)
		case Between:
			b := orm.NewCondition()
			b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
			b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
			cond = cond.AndCond(b)
		default:
			cond = cond.And(pathName, criteria.Value)
		}

		if this.Debug {
			logs.Debug("*********************************************************")
			logs.Debug("** set condition default %v ", pathName)
			logs.Debug("*********************************************************")
		}

		condition = condition.AndCond(cond)
	}

	return condition
}

func (this *Criteria) buildConditionsOr(criterias []*Criteria, condition *orm.Condition) *orm.Condition {

	for _, c := range criterias {

		cond := orm.NewCondition()

		for _, criteria := range c.criterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
			case In:
				cond = cond.Or(pathName, criteria.InValues)
			case Ne:
				cond = cond.OrNot(pathName, criteria.Value)
			case NotIn:
				cond = cond.OrNot(pathName, criteria.InValues)
			case IsNull:
				cond = cond.Or(pathName, true)
			case IsNotNull:
				cond = cond.Or(pathName, false)
			case Between:
				b := orm.NewCondition()
				b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
				b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
				cond = cond.OrCond(b)
			default:
				if criteria.ForceAnd {
					cond = cond.And(pathName, criteria.Value)
				} else {
					cond = cond.Or(pathName, criteria.Value)
				}
			}

			if this.Debug {
				logs.Debug("*********************************************************")
				logs.Debug("** set condition or %v ", pathName)
				logs.Debug("*********************************************************")
			}
		}

		if this.Session.HasTenant() && this.Session.HasFilterTenant(this.Result) && !this.Session.IgnoreTenantFilter {
			cond = cond.And("Tenant", this.Session.Tenant)
		}

		condition = condition.OrCond(cond)
	}

	return condition
}

func (this *Criteria) buildConditionsAnd(criterias []*Criteria, condition *orm.Condition) *orm.Condition {
	for _, c := range criterias {

		cond := orm.NewCondition()

		for _, criteria := range c.criterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
			case Ne, NotIn:
				cond = cond.AndNot(pathName, criteria.Value)
			case IsNull:
				cond = cond.And(pathName, true)
			case IsNotNull:
				cond = cond.And(pathName, false)
			case Between:
				cond = cond.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
				cond = cond.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
			default:
				cond = cond.And(pathName, criteria.Value)
			}
			if this.Debug {
				logs.Debug("*********************************************************")
				logs.Debug("** set condition and %v ", pathName)
				logs.Debug("*********************************************************")
			}
		}

		condition = condition.AndCond(cond)

	}

	return condition
}

func (this *Criteria) buildConditionsAndOr(criterias []*Criteria, condition *orm.Condition) *orm.Condition {

	for _, c := range criterias {

		cond := orm.NewCondition()

		for _, criteria := range c.criterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
			case Ne, NotIn:
				cond = cond.OrNot(pathName, criteria.Value)
			case IsNull:
				cond = cond.Or(pathName, true)
			case IsNotNull:
				cond = cond.Or(pathName, false)
			case Between:
				b := orm.NewCondition()
				b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
				b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
				cond = cond.OrCond(b)
			default:
				if criteria.ForceAnd {
					cond = cond.And(pathName, criteria.Value)
				} else {
					cond = cond.Or(pathName, criteria.Value)
				}
			}

			if this.Debug {
				logs.Debug("*********************************************************")
				logs.Debug("** set condition and or %v ", pathName)
				logs.Debug("*********************************************************")
			}

		}

		if this.Session.HasTenant() && this.Session.HasFilterTenant(this.Result) && !this.Session.IgnoreTenantFilter {
			cond = cond.And("Tenant", this.Session.Tenant)
		}

		condition = condition.AndCond(cond)

	}

	return condition
}

func (this *Criteria) buildConditionsAndOrAnd(criterias []*CriteriaSet, condition *orm.Condition) *orm.Condition {

	if len(this.criteriasAndOrAnd) > 0 {

		cond := orm.NewCondition()

		for _, criteriaSet := range criterias {

			for _, ct := range criteriaSet.Criterias {

				other := orm.NewCondition()

				for _, criteria := range ct.criterias {
					pathName := this.getPathName(criteria)

					switch criteria.Expression {
					case Ne, NotIn:
						other = other.AndNot(pathName, criteria.Value)
					case IsNull:
						other = other.And(pathName, true)
					case IsNotNull:
						other = other.And(pathName, false)
					case Between:
						b := orm.NewCondition()
						b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
						b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
						other = other.AndCond(b)
					default:
						other = other.And(pathName, criteria.Value)
					}

					if this.Debug {
						logs.Debug("*********************************************************")
						logs.Debug("** set condition and or %v ", pathName)
						logs.Debug("*********************************************************")
					}

				}

				cond = cond.OrCond(other)

			}

		}

		condition = condition.AndCond(cond)
	}

	return condition
}

func (this *Criteria) buildConditionsOrAnd(criterias []*Criteria, condition *orm.Condition) *orm.Condition {
	for _, c := range criterias {

		cond := orm.NewCondition()

		for _, criteria := range c.criterias {
			pathName := this.getPathName(criteria)

			switch criteria.Expression {
			case Ne, NotIn:
				cond = cond.AndNot(pathName, criteria.Value)
			case IsNull:
				cond = cond.And(pathName, true)
			case IsNotNull:
				cond = cond.And(pathName, false)
			case Between:
				b := orm.NewCondition()
				b = b.And(fmt.Sprintf("%v__gte", criteria.Path), criteria.Value)
				b = b.And(fmt.Sprintf("%v__lte", criteria.Path), criteria.Value2)
				cond = cond.AndCond(b)
			default:
				cond = cond.And(pathName, criteria.Value)
			}

			if this.Debug {
				logs.Debug("*********************************************************")
				logs.Debug("** set condition and or and %v ", pathName)
				logs.Debug("*********************************************************")
			}

		}

		if this.Session.HasTenant() && this.Session.HasFilterTenant(this.Result) && !this.Session.IgnoreTenantFilter {
			cond = cond.And("Tenant", this.Session.Tenant)
		}
		condition = condition.OrCond(cond)

	}

	return condition
}

func (this *Criteria) build(query orm.QuerySeter) orm.QuerySeter {

	condition := this.buildCriterias(this.criterias)

	condition = this.buildConditionsOr(this.criteriasOr, condition)

	condition = this.buildConditionsAnd(this.criteriasAnd, condition)

	condition = this.buildConditionsAndOr(this.criteriasAndOr, condition)

	condition = this.buildConditionsAndOrAnd(this.criteriasAndOrAnd, condition)

	condition = this.buildConditionsOrAnd(this.criteriasOrAnd, condition)

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
	case Like, NotLike:

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

	case IsNull, IsNotNull:
		pathName = fmt.Sprintf("%v__isnull", criteria.Path)
	//case IsNotNull:
	//	pathName = fmt.Sprintf("%v__isnull", criteria.Path)
	case Between:

	case In, NotIn:
		pathName = fmt.Sprintf("%v__in", criteria.Path)
		//case NotIn:
		//	pathName = fmt.Sprintf("%v__in", criteria.Path)

	}

	return pathName
}

func (this *Criteria) execute(resultType CriteriaResult) *Criteria {

	this.criteriaType = resultType

	defer func() {
		if this.tenantCopy != nil {
			this.Session.SetTenant(this.tenantCopy)
			this.tenantCopy = nil
		}
	}()

	if hook, ok := this.Result.(ModelHookBeforeCriteria); ok {
		hook.BeforeCriteria(this)
	}

	query := this.Query()

	if this.Limit > 0 {
		query = query.Limit(this.Limit).Offset(this.Offset)
	}

	if this.Session.HasTenant() && this.Session.HasFilterTenant(this.Result) {
		this.Eq("Tenant", this.Session.Tenant)
	}

	this.buildSearchPaths(this.searchPaths, this.searchValue)
	this.buildPage()
	query = this.build(query)

	if this.Distinct {
		query = query.Distinct()
	}

	switch resultType {

	case CriteriaAggregateOne:

		query = query.Aggregate(this.aggregate)

		if len(this.groupBy) > 0 {
			query = query.GroupBy(this.groupBy)
		}

		err := this.Session.ToOne(query, this.resultAggregate)

		if err != nil && err != orm.ErrNoRows && !strings.Contains(err.Error(), "repeat register") {
			this.SetError(err)
		} else {
			this.SetError(nil)
		}

		break

	case CriteriaAggregateList:

		query = query.Aggregate(this.aggregate)

		if len(this.groupBy) > 0 {
			query = query.GroupBy(this.groupBy)
		}

		err := this.Session.ToList(query, this.resultsAggregate)

		if err != nil && err != orm.ErrNoRows && !strings.Contains(err.Error(), "repeat register") {
			this.SetError(err)
		} else {
			this.SetError(nil)
		}

		this.Any = reflect.ValueOf(this.resultsAggregate).Elem().Len() > 0
		this.Empty = !this.Any

		break

	case CriteriaList:

		orders := []string{}
		for _, order := range this.orderBy {
			if order.Desc {
				orders = append(orders, fmt.Sprintf("-%v", order.Path))
			} else {
				orders = append(orders, fmt.Sprintf(order.Path))
			}
		}

		if len(orders) > 0 {
			query = query.OrderBy(orders...)
		}

		if len(this.RelatedSelList) > 0 {
			if len(this.RelatedSelList) == 1 && this.RelatedSelList[0] == "all" {
				query = query.RelatedSel()
			} else {
				for _, it := range this.RelatedSelList {
					query = query.RelatedSel(it)
				}
			}
		}

		if this.Results == nil {
			this.SetError(errors.New("Results can't be nil"))
			return this
		}

		err := this.Session.ToList(query, this.Results)

		this.SetError(err)

		this.Any = reflect.ValueOf(this.Results).Elem().Len() > 0
		this.Empty = !this.Any

		if hook, ok := this.Result.(ModelHookAfterList); ok {
			hook.AfterList(this.Results)
		}

	case CriteriaOne:

		orders := []string{}
		for _, order := range this.orderBy {
			if order.Desc {
				orders = append(orders, fmt.Sprintf("-%v", order.Path))
			} else {
				orders = append(orders, fmt.Sprintf(order.Path))
			}
		}

		if len(orders) > 0 {
			query = query.OrderBy(orders...)
		}

		if len(this.RelatedSelList) > 0 {
			if len(this.RelatedSelList) == 1 && this.RelatedSelList[0] == "all" {
				query = query.RelatedSel()
			} else {
				for _, it := range this.RelatedSelList {
					query = query.RelatedSel(it)
				}
			}
		}

		err := this.Session.ToOne(query, this.Result)

		if err != orm.ErrNoRows {
			this.SetError(err)
		} else {
			this.SetError(nil)
		}

		if !this.HasError {
			if hook, ok := this.Result.(ModelHookAfterLoad); ok {
				if next, err := hook.AfterLoad(this.Result); !next || err != nil {

					if err != nil {
						this.SetError(err)
						this.Result = nil
					}

					if !next {
						this.Result = nil
					}

				}
			}
		}

		if model, ok := this.Result.(Model); ok {
			this.Any = model.IsPersisted()
		}
		this.Empty = !this.Any

	case CriteriaCount, CriteriaExists:

		count, err := this.Session.ToCount(query)

		this.Count64 = count
		this.Count32 = int(count)

		this.Any = count > 0
		this.Empty = !this.Any

		this.SetError(err)

	case CriteriaDelete:

		count, err := this.Session.ExecuteDelete(query)

		this.Count64 = count
		this.Count32 = int(count)

		this.Any = count > 0
		this.Empty = !this.Any
		this.SetError(err)

	case CriteriaUpdate:

		count, err := this.Session.ExecuteUpdate(query, this.UpdateParams)

		this.Count64 = count
		this.Count32 = int(count)

		this.Any = count > 0
		this.Empty = !this.Any

		this.SetError(err)
	}

	return this

}

func (this *Criteria) SetError(err error) {
	if err != nil && this.Error == nil {
		this.HasError = true
		this.Error = errors.New(this.getErrorDescription(err))
	}

	this.query = nil
}

func (this *Criteria) getErrorDescription(err error) string {

	//entity := this.Result
	if model, ok := this.Result.(Model); ok {
		return fmt.Sprintf("Table: %v - Message: %v", model.TableName(), err)
	} else {
		return "entity does not implements of Model"
	}
}

func (this *Criteria) TryOneById(id int64) interface{} {
	this.Eq("Id", id)
	return this.TryOne()
}

func (this *Criteria) TryOne() interface{} {
	this.One()

	if this.HasError {
		return optional.NewFail(this.Error)
	}

	if this.Empty {
		return optional.NewNone()
	}

	return optional.NewSome(this.Result)
}

func (this *Criteria) TryList() interface{} {
	this.List()

	if this.HasError {
		return optional.NewFail(this.Error)
	}

	/*
		if this.Empty {
			return optional.NewEmpty()
		}
	*/

	val := reflect.ValueOf(this.Results)
	return optional.NewSome(val.Elem().Interface())
}
