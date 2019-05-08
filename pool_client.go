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
}

func (c *Client) Close() {
	c.over.closeClient(c)
}
