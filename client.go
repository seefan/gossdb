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
	if this != nil {
		if this.pool == nil { //连接池不存在，只关闭自己的连接
			this.Client.Close()
		} else {
			this.pool.closeClient(this)
		}
	}
}
