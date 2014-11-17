package gossdb

import (
	"github.com/ssdb/gossdb/ssdb"
	"time"
)

//可关闭连接
type Client struct {
	ssdb.Client
	pool     *Connectors //来源的连接池
	lastTime time.Time   //最后的更新时间
}

//关闭连接
func (this *Client) Close() {
	this.lastTime = time.Now()
	this.pool.pool <- this
}
