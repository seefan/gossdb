package gossdb

import (
	"encoding/json"
	"github.com/seefan/goerr"
	"github.com/seefan/to"
	"github.com/ssdb/gossdb/ssdb"
	"log"
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

//查询数据库大小
//
//  返回 re，返回数据库的估计大小, 以字节为单位. 如果服务器开启了压缩, 返回压缩后的大小.
//  返回 err，执行的错误
func (this *Client) DbSize() (re int, err error) {
	resp, err := this.Do("dbsize")
	if err != nil {
		return -1, err
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return strconv.Atoi(resp[1])
	}
	return -1, makeError(resp)
}

//返回服务器的信息.
//
//  返回 re，返回数据库的估计大小, 以字节为单位. 如果服务器开启了压缩, 返回压缩后的大小.
//  返回 err，执行的错误
func (this *Client) Info() (re []string, err error) {
	resp, err := this.Do("info")
	if err != nil {
		return nil, err
	}
	log.Println(resp)
	if len(resp) > 1 && resp[0] == "ok" {
		return resp[1:], nil
	}
	return nil, makeError(resp)
}

//对数据进行编码
func (this *Client) encoding(value interface{}, hasArray bool) string {
	switch t := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
		return to.String(t)
	case string: //byte==uint8
		return t
	case []byte:
		return string(t)
	case bool:
		if t {
			return "1"
		} else {
			return "0"
		}
	case nil:
		return ""
	case []bool, []string, []int, []int8, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64, []float32, []float64, []interface{}:
		if hasArray && this.pool.Encoding {
			if bs, err := json.Marshal(value); err == nil {
				return string(bs)
			}
		}
		return "can not support slice,please open the Encoding options"
	default:
		if this.pool.Encoding {
			if bs, err := json.Marshal(value); err == nil {
				return string(bs)
			}
		}
		return "not open Encoding options"
	}
}

//生成通过的错误信息，已经确定是有错误
func makeError(resp []string, errKey ...interface{}) error {
	if len(resp) < 1 {
		return goerr.New("ssdb respone error")
	}
	//正常返回的不存在不报错，如果要捕捉这个问题请使用exists
	if resp[0] == "not_found" {
		return nil
	}
	if len(errKey) > 0 {
		return goerr.New("access ssdb error, code is %v, parameter is %v", resp, errKey)
	} else {
		return goerr.New("access ssdb error, code is %v", resp)
	}
}
