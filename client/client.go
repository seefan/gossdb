package client

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/seefan/gossdb/ssdb_client"
)

const (
	OK       string = "ok"
	NotFound string = "not_found"
)

//可回收的连接，支持连接池。
//非协程安全，多协程请使用多个连接。
type Client struct {
	isOpen bool
	ssdb_client.SSDBClient
}

func NewClient(c *ssdb_client.SSDBClient) *Client {
	return &Client{
		SSDBClient: *c,
	}
}
func (c *Client) IsOpen() bool {
	return c.isOpen
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
	if len(resp) == 2 && resp[0] == OK {
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
	if len(resp) > 1 && resp[0] == OK {
		return resp[1:], nil
	}
	return nil, makeError(resp)
}

//生成通过的错误信息，已经确定是有错误
func makeError(resp []string, errKey ...interface{}) error {
	if len(resp) < 1 {
		return errors.New("ssdb respone error")
	}
	//正常返回的不存在不报错，如果要捕捉这个问题请使用exists
	if resp[0] == NotFound {
		return nil
	}
	if len(errKey) > 0 {
		return fmt.Errorf("access ssdb error, code is %v, parameter is %v", resp, errKey)
	} else {
		return fmt.Errorf("access ssdb error, code is %v", resp)
	}
}
