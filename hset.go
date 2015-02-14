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
		return goerr.NewError(err, "保存 %s/%s 值时出错", setName, key)
	}

	if len(resp) == 2 && resp[0] != "ok" {
		return goerr.New("保存set时出错", resp[1])
	}
	return nil
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
		return "", goerr.NewError(err, "获取 %s/%s 值时出错", setName, key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", nil
}
