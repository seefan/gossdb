package gossdb

import (
	"github.com/seefan/goerr"
)

//设置指定 key 的值内容
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  返回可能的错误
func (this *Client) Set(key string, val interface{}) error {
	resp, err := this.Client.Do("set", key, this.encoding(val, false))
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
