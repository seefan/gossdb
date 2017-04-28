package balance

import (
	"github.com/seefan/gossdb/client"
)

type IPoolProxy interface {
	Get() (client.IClient, error)
	Set(c client.IClient)
	Start() error
	Close()
	Info() string
}
