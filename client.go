package gossdb

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/seefan/goerr"
	"github.com/seefan/to"
	"github.com/ssdb/gossdb/ssdb"
)

//可回收的连接，支持连接池。
//非协程安全，多协程请使用多个连接。
type Client struct {
	ssdb.Client
	pool     *Connectors //来源的连接池
	lastTime time.Time   //最后的更新时间
	isOpen   bool        //是否已连接
	password string      //校验密码
}

//打开连接
func (this *Client) Start() error {
	if this.isOpen {
		return nil
	}
	//log.Println("set pwd", AuthPassword)
	this.password = AuthPassword
	db, err := ssdb.Connect(this.pool.cfg.Host, this.pool.cfg.Port)
	if err != nil {
		return err
	}
	this.lastTime = time.Now()
	this.isOpen = true
	this.Client = *db
	return nil
}

//关闭连接
func (this *Client) Close() {
	this.lastTime = time.Now()
	if this.pool == nil { //连接池不存在，只关闭自己的连接
		if this.isOpen {
			this.Client.Close()
			this.isOpen = false
		}
	} else {
		if this.isOpen {
			this.pool.closeClient(this)
		}
	}
}

//检查连接情况
//
//  返回 bool，如果可以正常查询数据库信息，就返回true，否则返回false
func (this *Client) Ping() bool {
	_, err := this.Info()
	return err == nil
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
	if len(resp) > 1 && resp[0] == "ok" {
		return resp[1:], nil
	}
	return nil, makeError(resp)
}

//对数据进行编码
func (this *Client) encoding(value interface{}, hasArray ...bool) string {
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
func (this *Client) Do(args ...interface{}) ([]string, error) {
	//log.Println("auth:", this.password, args)
	if this.password != "" {
		resp, err := this.Client.Do("auth", []string{this.password})
		if err != nil {
			this.Client.Close()
			this.isOpen = false
			return nil, goerr.NewError(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == "ok" {
			//验证成功
			this.password = ""
		} else {
			return nil, makeError(resp, "Authentication failed,password is wrong")
		}
	}
	resp, err := this.Client.Do(args...)
	if err != nil {
		this.Client.Close()
		this.isOpen = false
	}
	return resp, err
}

//配置密码, 之后将用于向服务器校验. 这个校验不是立即进行的, 而是等你执行第一条命令的时候才发给服务器. 注意, 密码是明文传输的!
//优先使用gossdb.AuthPassowrd配置，本方法仅用于临时修改
//  password 校验密码

func (this *Client) Auth(password string) {
	this.password = password
}
