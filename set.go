package gossdb

import (
	"log"
	//	"fmt"
	"github.com/seefan/goerr"
)

//设置指定 key 的值内容
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  ttl 可选，设置的过期时间，单位为秒
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Set(key string, val interface{}, ttl ...int) (err error) {
	var resp []string
	if len(ttl) > 0 {
		resp, err = this.Client.Do("setx", key, this.encoding(val, false), ttl[0])
	} else {
		resp, err = this.Client.Do("set", key, this.encoding(val, false))
	}
	if err != nil {
		return goerr.NewError(err, "Set %s error", key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}

//获取指定 key 的值内容
//
//  key 键值
//  返回 一个 Value,可以方便的向其它类型转换
//  返回 一个可能的错误，操作成功返回 nil
func (this *Client) Get(key string) (Value, error) {
	resp, err := this.Client.Do("get", key)
	if err != nil {
		return "", goerr.NewError(err, "Get %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//更新 key 对应的 value, 并返回更新前的旧的 value.
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  返回 一个 Value,可以方便的向其它类型转换.如果 key 不存在则返回 "", 否则返回 key 对应的值内容.
//  返回 一个可能的错误，操作成功返回 nil
func (this *Client) Getset(key string, val interface{}) (Value, error) {
	resp, err := this.Client.Do("getset", key, val)
	if err != nil {
		return "", goerr.NewError(err, "Getset %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//设置过期
//
//  key 要设置过期的 key
//  ttl 存活时间(秒)
//  返回 re，设置是否成功，如果当前 key 不存在返回 false
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Expire(key string, ttl int) (re bool, err error) {
	resp, err := this.Do("expire", key, ttl)
	if err != nil {
		return false, goerr.NewError(err, "Expire %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, key, ttl)
}

//查询指定 key 是否存在
//
//  key 要查询的 key
//  返回 re，如果当前 key 不存在返回 false
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Exists(key string) (re bool, err error) {
	resp, err := this.Do("exists", key)
	if err != nil {
		return false, goerr.NewError(err, "Exists %s error", key)
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, key)
}

//删除指定 key
//
//  key 要删除的 key
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Del(key string) error {
	resp, err := this.Do("del", key)
	if err != nil {
		return goerr.NewError(err, "Del %s error", key)
	}

	//response looks like this: [ok 1]
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}

//返回 key(只针对 KV 类型) 的存活时间.
//
//  key 要删除的 key
//  返回 ttl，key 的存活时间(秒), -1 表示没有设置存活时间.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Ttl(key string) (ttl int64, err error) {
	resp, err := this.Do("ttl", key)
	if err != nil {
		return -1, goerr.NewError(err, "Ttl %s error", key)
	}

	//response looks like this: [ok 1]
	if len(resp) > 0 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, key)
}

//使 key 对应的值增加 num. 参数 num 可以为负数.
//
//  key 键值
//  num 增加的值
//  返回 val，整数，增加 num 后的新值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Incr(key string, num int64) (val int64, err error) {

	resp, err := this.Client.Do("incr", key, num)

	if err != nil {
		return -1, goerr.NewError(err, "Incr %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, key)
}

//批量设置一批 key-value.
//
//  包含 key-value 的字典
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiSet(kvs map[string]interface{}) (err error) {

	args := []string{}
	for k, v := range kvs {
		args = append(args, k)
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Client.Do("multi_set", args)

	if err != nil {
		return goerr.NewError(err, "MultiSet %s error", kvs)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, kvs)
}

//批量获取一批 key 对应的值内容.
//
//  key，要获取的 key，可以为多个
//  返回 val，一个包含返回的 map
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiGet(key ...string) (val map[string]Value, err error) {
	if len(key) == 0 {
		return make(map[string]Value), nil
	}
	resp, err := this.Client.Do("multi_get", key)

	if err != nil {
		return nil, goerr.NewError(err, "MultiGet %s error", key)
	}
	log.Println("MultiGet", resp)
	size := len(resp)
	if size > 0 && resp[0] == "ok" {
		val = make(map[string]Value)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1])
		}
		return val, nil
	}
	return nil, makeError(resp, key)
}

//批量删除一批 key 和其对应的值内容.
//
//  key，要删除的 key，可以为多个
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiDel(key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	resp, err := this.Client.Do("multi_del", key)

	if err != nil {
		return goerr.NewError(err, "MultiDel %s error", key)
	}
	log.Println("MultiDel", resp)
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}
