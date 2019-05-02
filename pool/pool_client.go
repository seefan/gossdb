/*
@Time : 2019-04-30 20:51
@Author : seefan
@File : pool_client
@Software: gossdb
*/
package pool

import "github.com/seefan/gossdb/ssdb_client"

//缓存的连接
//
//内部使用
type PooledClient struct {
	//pos
	index int
	// The poolWait to which this element belongs.
	pool *Block
	//value
	Client *ssdb_client.SSDBClient
	//used status
	isUsed bool
	//last time
	lastTime int64
}
