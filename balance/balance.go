package balance

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/seefan/goerr"
	"github.com/seefan/gopool"
	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
)

type BalancePool struct {
	nm            *NodeManager
	pools         map[string]*gopool.Pool
	cfgList       []conf.Config
	NewSSDBClient func(host string, port int, password string) *client.SSDBClient
	nodeSize      int
}

func NewBalancePool(cfgList []conf.Config) *BalancePool {
	return &BalancePool{
		nm:      NewNodeManager(),
		pools:   make(map[string]*gopool.Pool),
		cfgList: cfgList,
	}
}
func (b *BalancePool) Get(args ...interface{}) (*gopool.PooledClient, string, error) {
	id := ""
	if b.nodeSize == 1 {
		id = "ssdb"
	} else {
		for i, s := range args {
			id += fmt.Sprint(s)
			if i > 1 {
				break
			}
		}
		if n, err := b.nm.GetNode(id); err == nil {
			id = n
		} else {
			return nil, "", err
		}
	}
	pc, err := b.pools[id].Get()
	if err != nil {
		return nil, "", err
	}
	return pc, id, nil
}
func (b *BalancePool) Set(c *gopool.PooledClient, id string) {
	b.pools[id].Set(c)
}
func (b *BalancePool) Start() error {
	if len(b.cfgList) == 1 {
		if db, err := b.loadSSDB(b.cfgList[0]); err == nil {
			b.pools["ssdb"] = db
			b.nodeSize = 1
		} else {
			return err
		}
	} else {
		for _, cfg := range b.cfgList {
			if err := b.loadNode(cfg); err != nil {
				log.Warnf("load ssdb node error", err.Error())
			} else {
				b.nodeSize += 1
			}
		}
	}
	if b.nodeSize == 0 {
		return goerr.New("ssdb node size is 0")
	}
	return nil
}
func (b *BalancePool) Close() {
	b.nm.Reset()
	for k := range b.pools {
		b.pools[k].Close()
	}
}

func (b *BalancePool) Info() string {
	return "not support"
}

func (b *BalancePool) loadNode(cfg conf.Config) error {
	id := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if p, err := b.loadSSDB(cfg); err != nil {
		return err
	} else {
		b.pools[id] = p
	}
	weight := cfg.Weight * 3
	if weight < 1 {
		weight = 3
	}
	b.nm.Append(&Node{
		ID: id,
	}, weight)
	return nil
}

func (b *BalancePool) loadSSDB(cfg conf.Config) (*gopool.Pool, error) {
	pool := new(gopool.Pool)
	pool.NewClient = func() gopool.IClient {
		return &client.SSDBClient{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Password: cfg.Password,
		}
	}
	pool.GetClientTimeout = cfg.GetClientTimeout
	pool.MaxPoolSize = cfg.MaxPoolSize
	pool.MinPoolSize = cfg.MinPoolSize
	pool.AcquireIncrement = cfg.AcquireIncrement
	pool.MaxWaitSize = cfg.MaxWaitSize
	pool.HealthSecond = cfg.HealthSecond
	return pool, pool.Start()
}
