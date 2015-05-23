package gossdb

import (
	"github.com/seefan/goerr"
	"log"
)

//设置 hashmap 中指定 key 对应的值内容.
//
//  setName hashmap 的名字
//  key hashmap 的 key
//  value key 的值
//  返回 err，执行的错误
func (this *Client) Hset(setName, key string, value interface{}) (err error) {
	resp, err := this.Do("hset", setName, key, this.encoding(value, false))
	if err != nil {
		return goerr.NewError(err, "Hset %s/%s error", setName, key)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, key)
}

//获取 hashmap 中指定 key 的值内容.
//
//  setName hashmap 的名字
//  key hashmap 的 key
//  返回 value key 的值
//  返回 err，执行的错误
func (this *Client) Hget(setName, key string) (value Value, err error) {
	resp, err := this.Do("hget", setName, key)
	if err != nil {
		return "", goerr.NewError(err, "Hget %s/%s error", setName, key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, setName, key)
}

//删除 hashmap 中的指定 key，不能通过返回值来判断被删除的 key 是否存在.
//
//  setName hashmap 的名字
//  key hashmap 的 key
//  返回 err，执行的错误
func (this *Client) Hdel(setName, key string) (err error) {
	resp, err := this.Do("hdel", setName, key)
	if err != nil {
		return goerr.NewError(err, "Hdel %s/%s error", setName, key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, key)
}

//判断指定的 key 是否存在于 hashmap 中.
//
//  setName hashmap 的名字
//  key hashmap 的 key
//  返回 re，如果当前 key 不存在返回 false
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Hexists(setName, key string) (re bool, err error) {
	resp, err := this.Do("hexists", setName, key)
	if err != nil {
		return false, goerr.NewError(err, "Hexists %s/%s error", setName, key)
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, setName, key)
}

//删除 hashmap 中的所有 key
//
//  setName hashmap 的名字
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Hclear(setName string) (err error) {
	resp, err := this.Do("hclear", setName)
	if err != nil {
		return goerr.NewError(err, "Hclear %s error", setName)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName)
}

func (this *Client) Hscan(setName string, keyStart, keyEnd string, limit int64) (map[string]Value, error) {

	resp, err := this.Client.Do("hscan", setName, keyStart, keyEnd, limit)

	if err != nil {
		return nil, goerr.NewError(err, "Hscan %s %s %s %v error", setName, keyStart, keyEnd, limit)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		re := make(map[string]Value)
		size := len(resp)
		for i := 1; i < size-1; i += 2 {
			re[resp[i]] = Value(resp[i+1])
		}
		return re, nil
	}
	return nil, makeError(resp, setName, keyStart, keyEnd, limit)
}

func (this *Client) MultiHset(setName string, kvs map[string]interface{}) (err error) {

	args := []string{}
	for k, v := range kvs {
		args = append(args, k)
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Client.Do("multi_hset", setName, args)

	if err != nil {
		return goerr.NewError(err, "MultiHset %s %s error", setName, kvs)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, kvs)
}

func (this *Client) MultiHget(setName string, key ...string) (val map[string]Value, err error) {
	if len(key) == 0 {
		return make(map[string]Value), nil
	}
	resp, err := this.Client.Do("multi_hget", setName, key)

	if err != nil {
		return nil, goerr.NewError(err, "MultiHget %s %s error", setName, key)
	}
	log.Println("multihget keys=", key)
	log.Println("MultiHget", resp)
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

func (this *Client) MultiHdel(setName string, key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	resp, err := this.Client.Do("multi_hdel", key)

	if err != nil {
		return goerr.NewError(err, "MultiHdel %s %s error", setName, key)
	}
	log.Println("MultiHdel", resp)
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}
