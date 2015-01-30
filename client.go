package gossdb

import (
	"errors"
	"github.com/ssdb/gossdb/ssdb"
	"log"
	//	"log"
	"strconv"
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

//设置过期
//key 要设置过期的 key
//ttl 存活时间(秒)
//返回 re，设置是否成功，如果当前 key 不存在返回 false
//返回 err，执行的错误
func (this *Client) Expire(key string, ttl int) (re bool, err error) {
	resp, err := this.Do("expire", key, ttl)
	if err != nil {
		return false, err
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, errors.New("当前 Key 不存在")
}

// 查询指定 key 是否存在
//key 要查询的 key
//返回 re，如果当前 key 不存在返回 false
//返回 err，执行的错误
func (this *Client) Exists(key string) (re bool, err error) {
	resp, err := this.Do("exists", key)
	if err != nil {
		return false, err
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, errors.New("执行过程中有错误")
}

//查询数据库大小
//返回 re，返回数据库的估计大小, 以字节为单位. 如果服务器开启了压缩, 返回压缩后的大小.
//返回 err，执行的错误
func (this *Client) DbSize() (re int, err error) {
	resp, err := this.Do("dbsize")
	if err != nil {
		return -1, err
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return strconv.Atoi(resp[1])
	}
	return -1, errors.New("执行过程中有错误")
}

//返回服务器的信息.
//返回 re，返回数据库的估计大小, 以字节为单位. 如果服务器开启了压缩, 返回压缩后的大小.
//返回 err，执行的错误
func (this *Client) Info() (re []string, err error) {
	resp, err := this.Do("info")
	if err != nil {
		return nil, err
	}
	log.Println(resp)
	if len(resp) > 1 && resp[0] == "ok" {
		return resp[1:], nil
	}
	return nil, errors.New("执行过程中有错误")
}
