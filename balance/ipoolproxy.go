package balance

import (
	"github.com/seefan/gopool"
)

type IPoolProxy interface {
	Get() (*gopool.PooledClient, error)
	Set(c *gopool.PooledClient)
}
