package ruslanparser

import (
	"github.com/coopernurse/gorp"
)

type Opts struct {
	Db         *gorp.DbMap
	SessionNum int
	GateAddr   string

	SearchQueryConsts map[string]interface{}
}

type Session struct {
	Id string

	Opts         *Opts
	CurrentQuery string
}

type Parser struct {
	Opts        Opts
	Sessions    []string
	SearchQueue []string
}

func NewParser(o Opts) *Parser {
	a := new(Parser)
	a.Opts = o
	return a
}

func Setup() {

}
