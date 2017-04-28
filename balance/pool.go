package balance

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/seefan/gopool"
	"github.com/seefan/gossdb/client"
	"github.com/seefan/gossdb/conf"
)

type Pool struct {
	nm        *NodeManager
	pools     map[string]*gopool.Pool
	NewClient func(host string, port int, password string) gopool.IClient
	cfgList   []conf.Config
	pool      *gopool.Pool
}

func NewBalanceManager(cfgList []conf.Config) *Pool {
	return &Pool{
		nm:      NewNodeManager(),
		pools:   make(map[string]*gopool.Pool),
		cfgList: cfgList,
	}
}
func (b *Pool) Start() error {
	if len(b.cfgList) == 1 {
		p, err := b.loadSSDB(b.cfgList[0])
		if err != nil {
			return err
		}
		b.pool = p
	} else {
		for _, cfg := range b.cfgList {
			if err := b.loadNode(cfg); err != nil {
				log.Warnf("load ssdb node error", err.Error())
			}
		}
	}
	return nil
}
func (b *Pool) Close() {
	b.nm.Reset()
	for k := range b.pools {
		b.pools[k].Close()
	}
}
func (b *Pool) Set(c client.IClient) {

}
func (b *Pool) Get() (client.IClient, error) {
	return nil, nil
}
func (b *Pool) Info() string {
	return "not support"
}

func (b *Pool) loadNode(cfg conf.Config) error {
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

func (b *Pool) loadSSDB(cfg conf.Config) (*gopool.Pool, error) {
	pool := new(gopool.Pool)
	pool.NewClient = func() gopool.IClient {
		return b.NewClient(cfg.Host, cfg.Port, cfg.Password)
	}
	pool.GetClientTimeout = cfg.GetClientTimeout
	pool.MaxPoolSize = cfg.MaxPoolSize
	pool.MinPoolSize = cfg.MinPoolSize
	pool.AcquireIncrement = cfg.AcquireIncrement
	pool.MaxWaitSize = cfg.MaxWaitSize
	pool.HealthSecond = cfg.HealthSecond
	return pool, pool.Start()
}
