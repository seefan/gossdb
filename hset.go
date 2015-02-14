package gossdb

import (
	"github.com/seefan/goerr"
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
		return goerr.NewError(err, " set hashmap %s/%s error", setName, key)
	}

	if len(resp) == 2 && resp[0] == "ok" {
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
		return "", goerr.NewError(err, "get hashmap %s/%s error", setName, key)
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
		return goerr.NewError(err, "del hashmap %s/%s error", setName, key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, key)
}
