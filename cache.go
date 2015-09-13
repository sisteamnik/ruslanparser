package ruslanparser

import (
	"github.com/fatih/color"
	"github.com/lunny/nodb"
	"github.com/lunny/nodb/config"
	"github.com/pquerna/ffjson/ffjson"
)

type Cacher interface {
	Set(*Cmd) error
	Get(*Cmd) error
}

type NodbCache struct {
	db *nodb.DB
}

func NewNodbCache(path string) (*NodbCache, error) {
	cfg := new(config.Config)
	cfg.DataDir = path
	dbs, e := nodb.Open(cfg)
	if e != nil {
		return nil, e
	}

	db, e := dbs.Select(0)
	if e != nil {
		return nil, e
	}

	c := &NodbCache{
		db: db,
	}
	return c, nil
}

func (c NodbCache) Set(cmd *Cmd) error {
	bts, e := ffjson.Marshal(cmd)
	if e != nil {
		return e
	}
	color.Green("set cache %s(%d)", cmd, cmd.ResultNum)
	return c.db.Set([]byte(cmd.String()), bts)
}

func (c NodbCache) Get(cmd *Cmd) error {
	bts, e := c.db.Get([]byte(cmd.String()))
	if e != nil {
		color.Red(e.Error())
		return e
	}
	var cm = Cmd{}
	e = ffjson.Unmarshal(bts, &cm)
	if e != nil {
		color.Red(e.Error())
		return e
	}
	cmd.Result = cm.Result
	cmd.ResultNum = cm.ResultNum
	color.Green("get cache %s(%d)", cmd, cmd.ResultNum)
	return nil
}
