package pool

import (
	"github.com/seefan/gossdb/client"
)

type Client struct {
	client.Client
	index int //连接池中的位置
	pool  *Pool
	over  *Connectors
	//连接是否正在使用，防止重复关闭
	used bool
}

func (c *Client) Close() {
	if c.Error == nil && c.used {
		c.over.closeClient(c)
	}
}
