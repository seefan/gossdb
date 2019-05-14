package pool

import (
	"github.com/seefan/gossdb/client"
)

//Client pooled client
type Client struct {
	client.Client
	//连接是否正在使用，防止重复关闭
	used bool
	//连接池中的位置
	index int
	pool  *Pool
	over  *Connectors
}

//Close put the client to Connectors
func (c *Client) Close() {
	if c.Error == nil && c.used {
		c.over.closeClient(c)
	}
}
