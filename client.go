package gossdb

import (
	"encoding/json"
	"github.com/seefan/goerr"
	"github.com/seefan/gopool"
	"github.com/seefan/to"
	"github.com/ssdb/gossdb/ssdb"
	//	"log"
	"strconv"
)

//可回收的连接，支持连接池。
//非协程安全，多协程请使用多个连接。
type Client struct {
	db *ssdb.Client
	gopool.Element
	pool     *Connectors //来源的连接池
	password string      //校验密码
	isOpen   bool
}

//创建一个新的连接
func NewClient(p *Connectors) *Client {
	return &Client{
		password: AuthPassword,
		pool:     p,
	}
}

//打开连接
func (c *Client) Start() error {
	c.password = AuthPassword
	db, err := ssdb.Connect(c.pool.cfg.Host, c.pool.cfg.Port)
	if err != nil {
		return err
	}
	c.db = db
	c.isOpen = true
	return nil
}

//关闭连接
func (c *Client) Close() error {
	if c.pool == nil { //连接池不存在，只关闭自己的连接
		if c.isOpen {
			c.db.Close()
		}
	} else {
		c.pool.closeClient(c)
	}
	return nil
}

//检查连接情况
//
//  返回 bool，如果可以正常查询数据库信息，就返回true，否则返回false
func (c *Client) Ping() bool {
	_, err := c.Info()
	return err == nil
}

//查询数据库大小
//
//  返回 re，返回数据库的估计大小, 以字节为单位. 如果服务器开启了压缩, 返回压缩后的大小.
//  返回 err，执行的错误
func (c *Client) DbSize() (re int, err error) {
	resp, err := c.Do("dbsize")
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
func (c *Client) Info() (re []string, err error) {
	resp, err := c.Do("info")
	if err != nil {
		return nil, err
	}
	if len(resp) > 1 && resp[0] == "ok" {
		return resp[1:], nil
	}
	return nil, makeError(resp)
}

//对数据进行编码
func (c *Client) encoding(value interface{}, hasArray ...bool) string {
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
		if len(hasArray) > 0 && hasArray[0] && Encoding {
			if bs, err := json.Marshal(value); err == nil {
				return string(bs)
			}
		}
		return "can not support slice,please open the Encoding options"
	default:
		if Encoding {
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

//通用调用方法，如果有需要在所有方法前执行的，可以在这里执行
func (c *Client) Do(args ...interface{}) ([]string, error) {

	if c.password != "" {
		resp, err := c.db.Do("auth", []string{c.password})
		if err != nil {
			c.db.Close()
			c.isOpen = false
			return nil, goerr.NewError(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == "ok" {
			//验证成功
			c.password = ""
		} else {
			return nil, makeError(resp, "Authentication failed,password is wrong")
		}
	}
	resp, err := c.db.Do(args...)
	if err != nil {
		c.db.Close()
		c.isOpen = false
	}
	return resp, err
}

//配置密码, 之后将用于向服务器校验. 这个校验不是立即进行的, 而是等你执行第一条命令的时候才发给服务器. 注意, 密码是明文传输的!
//优先使用gossdb.AuthPassowrd配置，本方法仅用于临时修改
//  password 校验密码

func (c *Client) Auth(password string) {
	c.password = password
}
