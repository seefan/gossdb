/*
@Time : 2019-05-06 20:33
@Author : seefan
@File : pool_client
@Software: gossdb
*/
package gossdb

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
