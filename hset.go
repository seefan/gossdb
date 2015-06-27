package gossdb

import "github.com/seefan/goerr"

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

//列出 hashmap 中处于区间 (key_start, key_end] 的 key-value 列表. ("", ""] 表示整个区间.
//
//  setName - hashmap 的名字.
//  keyStart - 返回的起始 key(不包含), 空字符串表示 -inf.
//  keyEnd - 返回的结束 key(包含), 空字符串表示 +inf.
//  limit - 最多返回这么多个元素.
//  返回包含 key-value 的关联字典.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Hscan(setName string, keyStart, keyEnd string, limit int64) (map[string]Value, error) {

	resp, err := this.Do("hscan", setName, keyStart, keyEnd, limit)

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

//批量设置 hashmap 中的 key-value.
//
//  setName - hashmap 的名字.
//  kvs - 包含 key-value 的关联数组 .
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) MultiHset(setName string, kvs map[string]interface{}) (err error) {

	args := []string{}
	for k, v := range kvs {
		args = append(args, k)
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Do("multi_hset", setName, args)

	if err != nil {
		return goerr.NewError(err, "MultiHset %s %s error", setName, kvs)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, kvs)
}

//批量获取 hashmap 中多个 key 对应的权重值.
//
//  setName - hashmap 的名字.
//  keys - 包含 key 的数组 .
//  返回 包含 key-value 的关联数组, 如果某个 key 不存在, 则它不会出现在返回数组中.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) MultiHget(setName string, key ...string) (val map[string]Value, err error) {
	if len(key) == 0 {
		return make(map[string]Value), nil
	}
	resp, err := this.Do("multi_hget", setName, key)

	if err != nil {
		return nil, goerr.NewError(err, "MultiHget %s %s error", setName, key)
	}
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

//批量获取 hashmap 中多个 key 对应的权重值.
//
//  setName - hashmap 的名字.
//  keys - 包含 key 的数组 .
//  返回 包含 key和value 的有序数组, 如果某个 key 不存在, 则它不会出现在返回数组中.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) MultiHgetSlice(setName string, key ...string) (keys []string, values []Value, err error) {
	if len(key) == 0 {
		return []string{}, []Value{}, nil
	}
	resp, err := this.Do("multi_hget", setName, key)

	if err != nil {
		return nil, nil, goerr.NewError(err, "MultiHgetSlice %s %s error", setName, key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		size := len(resp)
		keys := make([]string, 0, (size-1)/2)
		values := make([]Value, 0, (size-1)/2)

		for i := 1; i < size && i+1 < size; i += 2 {
			keys = append(keys, resp[i])
			values = append(values, Value(resp[i+1]))
		}
		return keys, values, nil
	}
	return nil, nil, makeError(resp, key)
}

//批量删除 hashmap 中的 key.
//
//  setName - hashmap 的名字.
//  keys - 包含 key 的数组.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) MultiHdel(setName string, key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	resp, err := this.Do("multi_hdel", key)

	if err != nil {
		return goerr.NewError(err, "MultiHdel %s %s error", setName, key)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}

//列出名字处于区间 (name_start, name_end] 的 hashmap. ("", ""] 表示整个区间.
//
//  keyStart - 返回的起始 key(不包含), 空字符串表示 -inf.
//  keyEnd - 返回的结束 key(包含), 空字符串表示 +inf.
//  limit - 最多返回这么多个元素.
//  返回 包含名字的数组
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Hlist(nameStart, nameEnd string, limit int64) ([]string, error) {
	resp, err := this.Do("hlist", nameStart, nameEnd, this.encoding(limit, false))
	if err != nil {
		return nil, goerr.NewError(err, "Hlist %s %s %v error", nameStart, nameEnd, limit)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		size := len(resp)
		keyList := make([]string, 0, size-1)

		for i := 1; i < size; i += 1 {
			keyList = append(keyList, resp[i])
		}
		return keyList, nil
	}
	return nil, makeError(resp, nameStart, nameEnd, limit)
}
