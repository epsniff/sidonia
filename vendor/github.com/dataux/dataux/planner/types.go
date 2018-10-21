package planner

import (
	"fmt"

	"github.com/lytics/dfa"

	"github.com/araddon/qlbridge/plan"
)

var (
	// States
	Starting    = dfa.State("starting")
	Running     = dfa.State("running")
	Resending   = dfa.State("resending")
	Finishing   = dfa.State("finishing")
	Exiting     = dfa.State("exiting")
	Terminating = dfa.State("terminating")
	// Letters
	Failure            = dfa.Letter("failure")
	SendFailure        = dfa.Letter("send-failure")
	SendSuccess        = dfa.Letter("send-success")
	FetchStateFailure  = dfa.Letter("fetch-state-failure")
	StoreStateFailure  = dfa.Letter("store-state-failure")
	EverybodyStarted   = dfa.Letter("everybody-started")
	EverybodyFinished  = dfa.Letter("everybody-finished")
	IndividualFinished = dfa.Letter("individual-finished")
	Exit               = dfa.Letter("exit")
)

type JobMaker func(ctx *plan.Context) (*ExecutorGrid, error)

type Flow string

func NewFlow(nr uint64) Flow {
	return Flow(fmt.Sprintf("sql-%v", nr))
}

func (f Flow) NewContextualName(name string) string {
	return fmt.Sprintf("%v-%v", f, name)
}

func (f Flow) Name() string {
	return string(f)
}

func (f Flow) String() string {
	return string(f)
}

type Conf struct {
	JobMaker       JobMaker
	SchemaLoader   plan.SchemaLoader
	SupressRecover bool
	NodeCt         int
	GridName       string
	Hostname       string
	EtcdServers    []string
	NatsServers    []string
}

func (c *Conf) Clone() *Conf {
	if c.JobMaker == nil {
		//panic("need job maker")
	}
	if c.SchemaLoader == nil {
		//panic("need SchemaLoader")
	}
	return &Conf{
		JobMaker:       c.JobMaker,
		SchemaLoader:   c.SchemaLoader,
		SupressRecover: c.SupressRecover,
		NodeCt:         c.NodeCt,
		GridName:       c.GridName,
		Hostname:       c.Hostname,
		EtcdServers:    c.EtcdServers,
		NatsServers:    c.NatsServers,
	}
}
