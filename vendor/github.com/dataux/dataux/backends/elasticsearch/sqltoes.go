package elasticsearch

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	u "github.com/araddon/gou"

	"github.com/araddon/qlbridge/exec"
	"github.com/araddon/qlbridge/expr"
	"github.com/araddon/qlbridge/lex"
	"github.com/araddon/qlbridge/plan"
	"github.com/araddon/qlbridge/rel"
	"github.com/araddon/qlbridge/schema"
	"github.com/araddon/qlbridge/value"
	"github.com/araddon/qlbridge/vm"
)

var (
	DefaultLimit = 1000

	// Implement Datasource interface that allows Mongo
	//  to fully implement a full select statement
	_ plan.SourcePlanner  = (*SqlToEs)(nil)
	_ exec.ExecutorSource = (*SqlToEs)(nil)
)

type esMap map[string]interface{}

// Sql To Elasticsearch Request Object
//   Map sql queries into Elasticsearch Json Requests
type SqlToEs struct {
	resp           *ResultReader
	p              *plan.Source
	tbl            *schema.Table
	sel            *rel.SqlSelect
	schema         *schema.SchemaSource
	ctx            *plan.Context
	partition      *schema.Partition // current partition for this request
	req            esMap             // Full request
	filter         esMap
	aggs           esMap
	groupby        esMap
	innergb        esMap   // InnerMost Group By
	sort           []esMap // sort
	hasMultiValue  bool    // Multi-Value vs Single-Value aggs
	hasSingleValue bool    // single value agg
	needsPolyFill  bool    // do we request that features be polyfilled?
	projections    map[string]string
}

func NewSqlToEs(table *schema.Table) *SqlToEs {
	return &SqlToEs{
		tbl:         table,
		schema:      table.SchemaSource,
		projections: make(map[string]string),
	}
}

func (m *SqlToEs) Close() error { return nil }

func (m *SqlToEs) Host() string {
	//u.Warnf("TODO:  replace hardcoded es host")
	return chooseBackend(m.schema)
}
func chooseBackend(schema *schema.SchemaSource) string {
	if len(schema.Conf.Nodes) == 0 {
		if len(schema.Conf.Hosts) > 0 {
			return schema.Conf.Hosts[0]
		}
	}
	for _, node := range schema.Conf.Nodes {
		// TODO:  implement real balancer
		return node.Address
	}
	return ""
}

func (m *SqlToEs) Columns() []string {
	return m.resp.Columns()
}

// WalkSourceSelect is used during planning phase, to create a plan (plan.Task)
//  or error, and to report back any poly-fill necessary
func (m *SqlToEs) WalkSourceSelect(planner plan.Planner, p *plan.Source) (plan.Task, error) {

	var err error
	m.p = p
	req := p.Stmt.Source
	m.sel = p.Stmt.Source
	if m.sel.Limit == 0 && p.Final {
		req.Limit = DefaultLimit
	} else if req.Limit == 0 {
		req.Limit = 1000
	}

	if req.Where != nil {
		m.filter = esMap{}
		_, err = m.WalkNode(req.Where.Expr, &m.filter)
		if err != nil {
			u.Warnf("Could Not evaluate Where Node %s %v", req.Where.Expr.String(), err)
			return nil, err
		}
	}

	err = m.WalkSelectList()
	if err != nil {
		u.Warnf("Could Not evaluate Columns/Aggs %s %v", req.Columns.String(), err)
		return nil, err
	}
	if len(req.GroupBy) > 0 {
		err = m.WalkGroupBy()
		if err != nil {
			u.Warnf("Could Not evaluate GroupBys %s %v", req.GroupBy.String(), err)
			return nil, err
		}
	}

	if len(m.aggs) > 0 && len(m.filter) > 0 {
		//u.Debugf("adding filter to aggs: %v", m.filter)
		m.aggs = esMap{"where": esMap{"aggs": m.aggs, "filter": m.filter}}
	}

	esReq := esMap{}
	//esPointer := esReq
	if len(m.groupby) > 0 {
		esReq = m.groupby
		esReq["size"] = 0
		req.Limit = 0
		//u.Infof("has group by: %v\ninner:%v   \naggs=%v", m.groupby, m.innergb, m.aggs)
		if len(m.aggs) > 0 {
			m.innergb["aggs"] = m.aggs
		}
	} else if len(m.aggs) > 0 {
		esReq = esMap{"aggs": m.aggs, "size": 0}
		req.Limit = 0
		//u.Debugf("setting limit: %v", req.Limit)
	} else if len(m.filter) > 0 {
		//u.Infof("in else: %v", esReq)
		esReq = esMap{"filter": m.filter}
	}
	//u.Debugf("OrderBy? %v  %#v", len(m.sel.OrderBy), m.sel.OrderBy)
	if len(m.sel.OrderBy) > 0 {
		m.sort = make([]esMap, len(m.sel.OrderBy))
		esReq["sort"] = m.sort
		for i, col := range m.sel.OrderBy {

			// We really need to look at any funcs?   walk this out
			switch col.Order {
			case "ASC", "DESC":
				m.sort[i] = esMap{col.As: esMap{"order": strings.ToLower(col.Order)}}
				u.Infof("sort col: %#v", col)
			default:
				// default sorder order = ?
				u.Warnf("not asc,desc???   %#v", col)
				m.sort[i] = esMap{col.As: esMap{"order": "asc"}}
			}
		}
	}

	p.Complete = true
	m.req = esReq
	return nil, nil
}

func (m *SqlToEs) WalkExecSource(p *plan.Source) (exec.Task, error) {

	if p.Stmt == nil {
		return nil, fmt.Errorf("Plan did not include Sql Statement?")
	}
	if p.Stmt.Source == nil {
		return nil, fmt.Errorf("Plan did not include Sql Select Statement?")
	}
	if m.p == nil {
		u.Debugf("custom? %v", p.Custom)

		m.p = p
		if p.Custom.Bool("poly_fill") {
			m.needsPolyFill = true
		}
		if partitionId := p.Custom.String("partition"); partitionId != "" {
			if p.Tbl.Partition != nil {
				for _, pt := range p.Tbl.Partition.Partitions {
					if pt.Id == partitionId {
						//u.Debugf("partition: %s   %#v", partitionId, pt)
						m.partition = pt
						if len(m.filter) == 0 {
							if pt.Left == "" {
								//m.filter = bson.M{p.Tbl.Partition.Keys[0]: bson.M{"$lt": pt.Right}}
							} else if pt.Right == "" {
								//m.filter = bson.M{p.Tbl.Partition.Keys[0]: bson.M{"$gte": pt.Left}}
							} else {

							}
						}
					}
				}
			}
		}
	}

	req := p.Stmt.Source

	// TODO:  hostpool
	qs := make(url.Values)
	qs.Set("size", strconv.Itoa(req.Limit))
	if len(m.projections) > 0 {
		fields := make([]string, 0)
		for field, _ := range m.projections {
			fields = append(fields, field)
		}
		// TODO: need to filter out non-leaf fields from above using
		//       schema, as es doesn't support filtering on non-leaf nodes
		//qs.Set("fields", strings.Join(fields, ","))
	}
	query := fmt.Sprintf("%s/%s/_search?%s", m.Host(), m.tbl.Name, qs.Encode())

	u.Infof("%v url=%v  filter=%v   \n\n%s", m.req, query, m.filter, u.JsonHelper(m.req).PrettyJson())
	jhResp, err := u.JsonHelperHttp("POST", query, m.req)
	if err != nil {
		u.Errorf("err %v", err)
		return nil, err
	}
	//u.Debugf("%s", jhResp.PrettyJson())

	if len(jhResp) == 0 {
		return nil, fmt.Errorf("No response, error fetching elasticsearch query")
	}

	resp := NewResultReader(m)
	m.resp = resp
	resp.Total = jhResp.Int("hits.total")
	resp.Aggs = jhResp.Helper("aggregations")
	if req.Where != nil {
		resp.Aggs = resp.Aggs.Helper("where")
		if len(resp.Aggs) > 0 {
			// Since we have a Where aggs, lets grab the count out of that????
			if docCt, ok := resp.Aggs.IntSafe("doc_count"); ok {
				resp.Total = docCt
			}
		}
	}

	resp.Docs = jhResp.Helpers("hits.hits")
	u.Debugf("p:%p resp %T  doc.ct = %v  cols:%v", resp, resp, len(resp.Docs), resp.Columns())
	p.Complete = true
	return resp, nil
}

// Aggregations from the <select_list>
//
//    SELECT <select_list> FROM ... WHERE
//
func (m *SqlToEs) WalkSelectList() error {

	m.aggs = esMap{}
	for i := len(m.sel.Columns) - 1; i >= 0; i-- {
		col := m.sel.Columns[i]
		//u.Debugf("i=%d of %d  %v %#v ", i, len(m.sel.Columns), col.Key(), col)
		if col.Expr != nil {
			switch curNode := col.Expr.(type) {
			// case *expr.NumberNode:
			// 	return nil, value.NewNumberValue(curNode.Float64), nil
			// case *expr.BinaryNode:
			// 	return m.walkBinary(curNode)
			// case *expr.TriNode: // Between
			// 	return m.walkTri(curNode)
			// case *expr.UnaryNode:
			// 	//return m.walkUnary(curNode)
			// 	u.Warnf("not implemented: %#v", curNode)
			case *expr.FuncNode:
				// All Func Nodes are Aggregates
				esm, err := m.WalkAggs(curNode)
				if err == nil && len(esm) > 0 {
					m.aggs[col.As] = esm
				} else if err != nil {
					u.Error(err)
					return err
				}
				//u.Debugf("esm: %v:%v", col.As, esm)
				//u.Debugf(curNode.String())
			// case *expr.MultiArgNode:
			// 	return m.walkMulti(curNode)
			// case *expr.StringNode:
			// 	return nil, value.NewStringValue(curNode.Text), nil
			case *expr.IdentityNode:
				s := curNode.String()
				m.projections[s] = s
			default:
				u.Debugf("likely a projection, not agg T:%T  %v", curNode, curNode)
				//panic("Unrecognized node type")
			}
		}

	}
	return nil
}

// Aggregations from the <select_list>
//
//    WHERE .. GROUP BY x,y,z
//
func (m *SqlToEs) WalkGroupBy() error {

	for i, col := range m.sel.GroupBy {
		if col.Expr != nil {
			//u.Infof("Walk group by %s  %T", col.Expr.String(), col.Expr)
			switch col.Expr.(type) {
			case *expr.IdentityNode, *expr.FuncNode:
				esm := esMap{}
				_, err := m.WalkNode(col.Expr, &esm)
				//fld := strings.Replace(expr.FindIdentityField(col.Expr), ".", "", -1)
				//u.Infof("gb: %s  %s", fld, u.JsonHelper(esm).PrettyJson())
				if err == nil {
					if len(m.innergb) > 0 {
						esm["aggs"] = esMap{fmt.Sprintf("group_by_%d", i): m.innergb}
						// esm["aggs"] = esMap{"group_by_" + fld: m.innergb}
					} else {
						esm = esm
					}
					m.innergb = esm
					//u.Infof("esm: %v", esm)
				} else {
					u.Error(err)
					return err
				}

			}
		}
	}

	m.groupby = esMap{"aggs": esMap{"group_by": m.innergb}}
	return nil
}

// WalkAggs() aggregate expressions when used ast part of <select_list>
//  - For Aggregates (functions) it builds aggs
//  - For Projectsion (non-functions) it does nothing, that will be done later during projection
func (m *SqlToEs) WalkAggs(cur expr.Node) (q esMap, _ error) {
	switch curNode := cur.(type) {
	// case *expr.NumberNode:
	// 	return nil, value.NewNumberValue(curNode.Float64), nil
	// case *expr.BinaryNode:
	// 	return m.walkBinary(curNode)
	// case *expr.TriNode: // Between
	// 	return m.walkTri(curNode)
	// case *expr.UnaryNode:
	// 	//return m.walkUnary(curNode)
	// 	u.Warnf("not implemented: %#v", curNode)
	case *expr.FuncNode:
		return m.walkAggFunc(curNode)
	// case *expr.MultiArgNode:
	// 	return m.walkMulti(curNode)
	// case *expr.IdentityNode:
	// 	return nil, value.NewStringValue(curNode.Text), nil
	// case *expr.StringNode:
	// 	return nil, value.NewStringValue(curNode.Text), nil
	default:
		u.Debugf("likely a projection, not agg T:%T  %v", cur, cur)
		//panic("Unrecognized node type")
	}
	// if cur.Negate {
	// 	q = esMap{"not": q}
	// }
	return q, nil
}

// Walk() an expression, and its AND/OR/() logic to create an appropriately
//  nested json document for elasticsearch queries
//
//  TODO:  think we need to separate Value Nodes from those that return es types?
func (m *SqlToEs) WalkNode(cur expr.Node, q *esMap) (value.Value, error) {
	//u.Debugf("walkFilter: %#v", cur)
	switch curNode := cur.(type) {
	case *expr.NumberNode, *expr.StringNode:
		nodeVal, ok := vm.Eval(nil, cur)
		if !ok {
			u.Warnf("not ok %v", cur)
			return nil, fmt.Errorf("could not evaluate: %v", cur.String())
		}
		u.Infof("nodeval? %v", nodeVal)
		return nodeVal, nil
		// What do we do here?
	case *expr.BinaryNode:
		return m.walkFilterBinary(curNode, q)
	case *expr.TriNode: // Between
		return m.walkFilterTri(curNode, q)
	case *expr.UnaryNode:
		//return m.walkUnary(curNode)
		u.Warnf("not implemented: %#v", curNode)
		return nil, fmt.Errorf("Not implemented urnary function: %v", curNode.String())
	case *expr.FuncNode:
		return m.walkFilterFunc(curNode, q)
	case *expr.IdentityNode:
		u.Warnf("wat????   %v", curNode.String())
		// "errors" :   { "term" : { "body" : "error"   }},
		//"terms": {        "field": "field1"      }
		*q = esMap{"terms": esMap{"field": curNode.String()}}
		return value.NewStringValue(curNode.String()), nil
	case *expr.ArrayNode:
		u.Warnf("not implemented arraynode")
		return nil, fmt.Errorf("not implemented arraynode")
	default:
		u.Errorf("unrecognized T:%T  %v", cur, cur)
		panic("Unrecognized node type")
	}
	// if cur.Negate {
	// 	q = esMap{"not": q}
	// }
	return nil, nil
}

func (m *SqlToEs) walkFilterTri(node *expr.TriNode, q *esMap) (value.Value, error) {

	arg1val, aok := vm.Eval(nil, node.Args[0])
	//u.Debugf("arg1? %v  ok?%v", arg1val, aok)
	arg2val, bok := vm.Eval(nil, node.Args[1])
	arg3val, cok := vm.Eval(nil, node.Args[2])
	//u.Debugf("walkTri: %v  %v %v %v", node, arg1val, arg2val, arg3val)
	if !aok || !bok || !cok {
		return nil, fmt.Errorf("Could not evaluate args: %v", node.String())
	}
	//u.Debugf("walkTri: %v  %v %v %v", node, arg1val, arg2val, arg3val)
	switch node.Operator.T {
	case lex.TokenBetween:
		*q = esMap{"range": esMap{arg1val.ToString(): esMap{"gte": arg2val.ToString(), "lte": arg3val.ToString()}}}
	default:
		u.Warnf("not implemented ")
	}
	if q != nil {
		return nil, nil
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *SqlToEs) walkFilterBinary(node *expr.BinaryNode, q *esMap) (value.Value, error) {

	// var lhval, rhval interface{}
	// switch curNode := cur.(type) {
	// case *expr.BinaryNode, *expr.TriNode, *expr.UnaryNode, *expr.MultiArgNode, *expr.FuncNode:
	// 	u.Warnf("not implemented: %#v", curNode)
	// 	panic("not implemented")
	// case *expr.IdentityNode, *expr.StringNode, *expr.NumberNode:
	// 	u.Infof("node? %v", curNode)
	// default:
	// 	u.Errorf("unrecognized T:%T  %v", cur, cur)
	// 	panic("Unrecognized node type")
	// }
	lhval, lhok := vm.Eval(nil, node.Args[0])
	rhval, rhok := vm.Eval(nil, node.Args[1])
	if !lhok || !rhok {
		u.Warnf("not ok: %v  l:%v  r:%v", node, lhval, rhval)
		return nil, fmt.Errorf("could not evaluate: %v", node.String())
	}
	//u.Debugf("walkBinary: op:%q  node:%v  l:%v  r:%v  %T  %T", node.Operator.V, node, lhval, rhval, lhval, rhval)
	switch node.Operator.T {
	case lex.TokenLogicAnd:
		// this doesn't yet implement x AND y AND z
		lhq, rhq := esMap{}, esMap{}
		_, err := m.WalkNode(node.Args[0], &lhq)
		_, err2 := m.WalkNode(node.Args[1], &rhq)
		if err != nil || err2 != nil {
			u.Errorf("could not get children nodes? %v %v %v", err, err2, node)
			return nil, fmt.Errorf("could not evaluate: %v", node.String())
		}
		*q = esMap{"and": []esMap{lhq, rhq}}
	case lex.TokenEqual, lex.TokenEqualEqual:
		//q = esMap{"terms": esMap{lhs: rhs}}
		if lhval != nil && rhval != nil {
			u.Debugf("ToString: %q", lhval.ToString())
			*q = esMap{"term": esMap{lhval.ToString(): rhval.ToString()}}
			return nil, nil
		}
		if lhval != nil || rhval != nil {
			u.Infof("has stuff?  %v", node.String())
		}
	case lex.TokenLE:
		*q = esMap{"range": esMap{lhval.ToString(): esMap{"lte": rhval.ToString()}}}
	case lex.TokenLT:
		*q = esMap{"range": esMap{lhval.ToString(): esMap{"lt": rhval.ToString()}}}
	case lex.TokenGE:
		*q = esMap{"range": esMap{lhval.ToString(): esMap{"gte": rhval.ToString()}}}
	case lex.TokenGT:
		*q = esMap{"range": esMap{lhval.ToString(): esMap{"gt": rhval.ToString()}}}
	case lex.TokenLike:
		//q = esMap{"terms": esMap{lhs: rhs}}
		switch val := lhval.ToString(); strings.ToLower(val) {
		case "all", "query", "any":
			//q = esMap{"query": esMap{"query_string": esMap{"query": rhval.ToString()}}}
			*q = esMap{"query": esMap{"query_string": esMap{"query": rhval.ToString()}}}
		default:
			if lhval != nil && rhval != nil {
				rhs := rhval.ToString()
				hasWildcard := strings.Contains(rhs, "%")
				if hasWildcard {
					rhs = strings.Replace(rhs, "%", "*", -1)
				} else {
					hasWildcard = strings.Contains(rhs, "*")
				}
				hasLogic := strings.Contains(rhs, " AND")
				if !hasLogic {
					hasLogic = strings.Contains(rhs, " OR")
				}
				if hasLogic || hasWildcard {
					*q = esMap{"query": esMap{"query_string": esMap{"query": rhs}}}
				} else {
					*q = esMap{"query": esMap{"match": esMap{lhval.ToString(): rhs}}}
				}

				return nil, nil
			}
		}

	case lex.TokenIN:
		switch vt := rhval.(type) {
		case value.SliceValue:
			// year IN (1990,1992)  =>
			//   		"filter" : {  "terms" : { "year" : [1990,1992]}   }
			*q = esMap{"terms": esMap{lhval.ToString(): vt.Values()}}
		default:
			u.Warnf("not implemented type %#v", rhval)
		}
	default:
		u.Warnf("not implemented: %v", node.Operator)
	}
	if q != nil {
		return nil, nil
	}
	return nil, fmt.Errorf("not implemented %v", node.String())
}

// Take an expression func, ensure we don't do runtime-checking (as the function)
//   doesn't really exist, then map that function to an ES Filter
//
//    exists, missing, prefix, term
//
func (m *SqlToEs) walkFilterFunc(node *expr.FuncNode, q *esMap) (value.Value, error) {
	switch funcName := strings.ToLower(node.Name); funcName {
	case "exists", "missing", "prefix", "term":
		if len(node.Args) != 1 {
			return nil, fmt.Errorf("Invalid func: %s", node.String())
		}
		fieldName := ""
		switch nt := node.Args[0].(type) {
		case *expr.IdentityNode:
			fieldName = nt.String()
		default:
			val, ok := eval(node.Args[0])
			if !ok {
				u.Errorf("Must be valid: %v", node.String())
			} else {
				fieldName = val.ToString()
			}
		}

		//  "exists" : { "field" : "user" }
		// "missing" : { "field" : "user" }
		*q = esMap{funcName: esMap{"field": fieldName}}
	default:
		u.Warnf("not implemented %s", node.String())
	}
	if q != nil {
		return nil, nil
	}
	return nil, fmt.Errorf("not implemented %v", node.String())
}

// Take an expression func, ensure we don't do runtime-checking (as the function)
//   doesn't really exist, then map that function to an ES aggregation
//
//    min, max, avg, sum, cardinality, terms
//
//   Single Value Aggregates:
//       min, max, avg, sum, cardinality, count
//
//  MultiValue aggregats:
//      terms, ??
//
func (m *SqlToEs) walkAggFunc(node *expr.FuncNode) (q esMap, _ error) {
	switch funcName := strings.ToLower(node.Name); funcName {
	case "max", "min", "avg", "sum", "cardinality":
		m.hasSingleValue = true
		if len(node.Args) != 1 {
			return nil, fmt.Errorf("Invalid func")
		}
		val, ok := eval(node.Args[0])
		if !ok {
			u.Errorf("Must be valid: %v", node.String())
		}
		// "min_price" : { "min" : { "field" : "price" } }
		q = esMap{funcName: esMap{"field": val.ToString()}}
	case "terms":
		m.hasMultiValue = true
		// "products" : { "terms" : {"field" : "product", "size" : 5 }}

		if len(node.Args) == 0 || len(node.Args) > 2 {
			return nil, fmt.Errorf("Invalid terms function terms(field,10) OR terms(field)")
		}
		val, ok := eval(node.Args[0])
		if !ok {
			u.Errorf("Must be valid: %v", node.String())
		}
		if len(node.Args) >= 2 {
			size, ok := vm.Eval(nil, node.Args[1])
			if !ok {
				u.Errorf("Must be valid size: %v", node.Args[1].String())
			}
			// "products" : { "terms" : {"field" : "product", "size" : 5 }}
			q = esMap{funcName: esMap{"field": val.ToString(), "size": size.Value()}}
		} else {

			q = esMap{funcName: esMap{"field": val.ToString()}}
		}

	case "count":
		m.hasSingleValue = true
		u.Debugf("how do we want to use count(*)?  hit.hits?   or exists()?")
		val, ok := eval(node.Args[0])
		if !ok {
			u.Errorf("Must be valid: %v", node.String())
			return nil, fmt.Errorf("Invalid argument: %v", node.String())
		}
		if val.ToString() == "*" {
			return nil, nil
		} else {
			return esMap{"exists": esMap{"field": val.ToString()}}, nil
		}

	default:
		u.Warnf("not implemented ")
	}
	u.Debugf("func:  %v", q)
	if q != nil {
		return q, nil
	}
	return nil, fmt.Errorf("not implemented")
}

func eval(cur expr.Node) (value.Value, bool) {
	switch curNode := cur.(type) {
	case *expr.IdentityNode:
		return value.NewStringValue(curNode.Text), true
	case *expr.StringNode:
		return value.NewStringValue(curNode.Text), true
	default:
		u.Errorf("unrecognized T:%T  %v", cur, cur)
	}
	return value.NilValueVal, false
}
