package router

import (
	"fmt"
	"strings"

	u "github.com/araddon/gou"

	"github.com/araddon/qlbridge/schema"
	"github.com/dataux/dataux/models"
)

var _ = u.EMPTY

type Rule struct {
	DB    string
	Table string
	Key   string
	Type  string
	Nodes []string
	Shard Shard
}

func (r *Rule) FindNode(key interface{}) string {
	i := r.Shard.FindForKey(key)
	return r.Nodes[i]
}

func (r *Rule) FindNodeIndex(key interface{}) int {
	return r.Shard.FindForKey(key)
}

func (r *Rule) String() string {
	return fmt.Sprintf("%s.%s?key=%v&shard=%s&nodes=%s",
		r.DB, r.Table, r.Key, r.Type, strings.Join(r.Nodes, ", "))
}

func NewDefaultRule(db string, node string) *Rule {
	var r *Rule = &Rule{
		DB:    db,
		Type:  DefaultRuleType,
		Nodes: []string{node},
		Shard: new(DefaultShard),
	}
	return r
}

func (r *Router) GetRule(table string) *Rule {
	rule := r.Rules[table]
	if rule == nil {
		return r.DefaultRule
	} else {
		return rule
	}
}

type Router struct {
	DB          string
	Rules       map[string]*Rule //key is <table name>
	DefaultRule *Rule
	nodes       []string //just for human visiblity
}

func NewRouter(conf *models.Config, schema *schema.Schema) (*Router, error) {

	panic("not implemented nodes")

	if !includeNodeConfig(conf.Nodes, conf.Rules.Default) {
		return nil, fmt.Errorf("default node[%s] not in the nodes list.",
			conf.Rules.Default)
	}

	rt := new(Router)
	rt.DB = schema.Name
	//rt.nodes = conf.Nodes
	rt.Rules = make(map[string]*Rule, len(conf.Rules.ShardRule))
	rt.DefaultRule = NewDefaultRule(rt.DB, conf.Rules.Default)

	for _, shard := range conf.Rules.ShardRule {
		u.Infof("shard: %v", shard)
		rc := &RuleConfig{shard}
		for _, node := range shard.Nodes {
			if !includeNode(rt.nodes, node) {
				return nil, fmt.Errorf("shard table[%s] node[%s] not in the schema.nodes list:[%s].",
					shard.Table, node, strings.Join(shard.Nodes, ","))
			}
		}
		rule, err := rc.ParseRule(rt.DB)
		if err != nil {
			return nil, err
		}

		if rule.Type == DefaultRuleType {
			return nil, fmt.Errorf("[default-rule] duplicate, must only one.")
		} else {
			if _, ok := rt.Rules[rule.Table]; ok {
				return nil, fmt.Errorf("table %s rule in %s duplicate", rule.Table, rule.DB)
			}
			rt.Rules[rule.Table] = rule
		}
	}
	return rt, nil
}

func includeNode(nodes []string, node string) bool {
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}

func includeNodeConfig(nodes []*schema.ConfigNode, node string) bool {
	for _, n := range nodes {
		if n.Name == node {
			return true
		}
	}
	return false
}
