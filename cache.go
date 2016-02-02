package ruslanparser

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/lunny/nodb"
	"github.com/lunny/nodb/config"
	"github.com/pquerna/ffjson/ffjson"
	"time"
)

type Cacher interface {
	Set(*Cmd) error
	Get(*Cmd) error
	SetLifeTime(time.Duration)
}

type NodbCache struct {
	db       *nodb.DB
	lifeTime time.Duration
}

type CachedCmd struct {
	*Cmd
	Created time.Time
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
		db:       db,
		lifeTime: 60 * time.Minute,
	}
	return c, nil
}

func (c NodbCache) Set(cmd *Cmd) error {
	ccmd := &CachedCmd{Cmd: cmd, Created: time.Now()}
	bts, e := ffjson.Marshal(ccmd)
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
	var cm = CachedCmd{}
	e = ffjson.Unmarshal(bts, &cm)
	if e != nil {
		color.Red(e.Error())
		return e
	}
	if time.Since(cm.Created) > c.lifeTime {
		return fmt.Errorf("not cached")
	}
	cmd.Result = cm.Result
	cmd.ResultNum = cm.ResultNum
	color.Green("get cache %s(%d)", cmd, cmd.ResultNum)
	return nil
}

func (c NodbCache) SetLifeTime(dur time.Duration) {
	c.lifeTime = dur
}
