package gossdb

import (
	"github.com/seefan/goerr"
)

//设置指定 key 的值内容
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  ttl 可选，设置的过期时间，单位为秒
//  返回可能的错误
func (this *Client) Set(key string, val interface{}, ttl ...int) (err error) {
	var resp []string
	if len(ttl) > 0 {
		resp, err = this.Client.Do("setx", key, this.encoding(val, false), ttl[0])
	} else {
		resp, err = this.Client.Do("set", key, this.encoding(val, false))
	}
	if err != nil {
		return goerr.NewError(err, "设置 %s 值时出错", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return nil
	}
	return goerr.New("设置 %s 值时出错，代码为：%s", key, resp[0])
}

//获取指定 key 的值内容
//
//  key 键值
//  返回一个 Value,可以方便的向其它类型转换
//  返回一个可能的错误
func (this *Client) Get(key string) (Value, error) {
	resp, err := this.Client.Do("get", key)
	if err != nil {
		return "", goerr.NewError(err, "获取 %s 值时出错", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", goerr.New("指定键值 %s 不存在", key)
}

//设置过期
//
//  key 要设置过期的 key
//  ttl 存活时间(秒)
//  返回 re，设置是否成功，如果当前 key 不存在返回 false
//  返回 err，执行的错误
func (this *Client) Expire(key string, ttl int) (re bool, err error) {
	resp, err := this.Do("expire", key, ttl)
	if err != nil {
		return false, err
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, goerr.New("执行过程中有错误，错误代码:%v", resp)
}

//查询指定 key 是否存在
//
//  key 要查询的 key
//  返回 re，如果当前 key 不存在返回 false
//  返回 err，执行的错误
func (this *Client) Exists(key string) (re bool, err error) {
	resp, err := this.Do("exists", key)
	if err != nil {
		return false, err
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, goerr.New("执行过程中有错误，错误代码:%v", resp)
}
