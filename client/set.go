package client

import (
	"github.com/seefan/goerr"
)

//Set 设置指定 key 的值内容
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  ttl 可选，设置的过期时间，单位为秒
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Set(key string, val interface{}, ttl ...int64) (err error) {
	var resp []string
	if len(ttl) > 0 {
		resp, err = c.Do("setx", key, val, ttl[0])
	} else {
		resp, err = c.Do("set", key, val)
	}
	if err != nil {
		return goerr.Errorf(err, "Set %s error", key)
	}
	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, key)
}

//SetNX 当 key 不存在时, 设置指定 key 的值内容. 如果已存在, 则不设置.
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  返回 err，可能的错误，操作成功返回 nil
//  返回 val 1: value 已经设置, 0: key 已经存在, 不更新.
func (c *Client) SetNX(key string, val interface{}) (Value, error) {
	resp, err := c.Do("setnx", key, val)

	if err != nil {
		return "", goerr.Errorf(err, "Setnx %s error", key)
	}
	if len(resp) > 0 && resp[0] == oK {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//Get 获取指定 key 的值内容
//
//  key 键值
//  返回 一个 Value,可以方便的向其它类型转换
//  返回 一个可能的错误，操作成功返回 nil
func (c *Client) Get(key string) (Value, error) {
	resp, err := c.Do("get", key)
	if err != nil {
		return "", goerr.Errorf(err, "Get %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//GetSet 更新 key 对应的 value, 并返回更新前的旧的 value.
//
//  key 键值
//  val 存贮的 value 值,val只支持基本的类型，如果要支持复杂的类型，需要开启连接池的 Encoding 选项
//  返回 一个 Value,可以方便的向其它类型转换.如果 key 不存在则返回 "", 否则返回 key 对应的值内容.
//  返回 一个可能的错误，操作成功返回 nil
func (c *Client) GetSet(key string, val interface{}) (Value, error) {
	resp, err := c.Do("getset", key, val)
	if err != nil {
		return "", goerr.Errorf(err, "Getset %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//Expire 设置过期
//
//  key 要设置过期的 key
//  ttl 存活时间(秒)
//  返回 re，设置是否成功，如果当前 key 不存在返回 false
//  返回 err，执行的错误，操作成功返回 nil
func (c *Client) Expire(key string, ttl int64) (re bool, err error) {
	resp, err := c.Do("expire", key, ttl)
	if err != nil {
		return false, goerr.Errorf(err, "Expire %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, key, ttl)
}

//Exists 查询指定 key 是否存在
//
//  key 要查询的 key
//  返回 re，如果当前 key 不存在返回 false
//  返回 err，执行的错误，操作成功返回 nil
func (c *Client) Exists(key string) (re bool, err error) {
	resp, err := c.Do("exists", key)
	if err != nil {
		return false, goerr.Errorf(err, "Exists %s error", key)
	}

	if len(resp) == 2 && resp[0] == oK {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, key)
}

//Del 删除指定 key
//
//  key 要删除的 key
//  返回 err，执行的错误，操作成功返回 nil
func (c *Client) Del(key string) error {
	resp, err := c.Do("del", key)
	if err != nil {
		return goerr.Errorf(err, "Del %s error", key)
	}

	//response looks like s: [ok 1]
	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, key)
}

//TTL 返回 key(只针对 KV 类型) 的存活时间.
//
//  key 要删除的 key
//  返回 ttl，key 的存活时间(秒), -1 表示没有设置存活时间.
//  返回 err，执行的错误，操作成功返回 nil
func (c *Client) TTL(key string) (ttl int64, err error) {
	resp, err := c.Do("ttl", key)
	if err != nil {
		return -1, goerr.Errorf(err, "Ttl %s error", key)
	}

	//response looks like s: [ok 1]
	if len(resp) > 0 && resp[0] == oK {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, key)
}

//Incr 使 key 对应的值增加 num. 参数 num 可以为负数.
//
//  key 键值
//  num 增加的值
//  返回 val，整数，增加 num 后的新值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Incr(key string, num int64) (val int64, err error) {

	resp, err := c.Do("incr", key, num)

	if err != nil {
		return -1, goerr.Errorf(err, "Incr %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, key)
}

//MultiSet 批量设置一批 key-value.
//
//  包含 key-value 的字典
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiSet(kvs map[string]interface{}) (err error) {

	args := []interface{}{"multi_set"}

	for k, v := range kvs {
		args = append(args, k)
		args = append(args, v)
	}
	resp, err := c.Do(args...)

	if err != nil {
		return goerr.Errorf(err, "MultiSet %s error", kvs)
	}

	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, kvs)
}

//MultiGet 批量获取一批 key 对应的值内容.
//
//  key，要获取的 key，可以为多个
//  返回 val，一个包含返回的 map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiGet(key ...string) (val map[string]Value, err error) {
	if len(key) == 0 {
		return make(map[string]Value), nil
	}
	data := []interface{}{"multi_get"}
	for _, k := range key {
		data = append(data, k)
	}
	resp, err := c.Do(data...)

	if err != nil {
		return nil, goerr.Errorf(err, "MultiGet %s error", key)
	}

	size := len(resp)
	if size > 0 && resp[0] == oK {
		val = make(map[string]Value)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1])
		}
		return val, nil
	}
	return nil, makeError(resp, key)
}

//MultiGetSlice 批量获取一批 key 对应的值内容.
//
//  key，要获取的 key，可以为多个
//  返回 keys和value分片
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiGetSlice(key ...string) (keys []string, values []Value, err error) {
	if len(key) == 0 {
		return []string{}, []Value{}, nil
	}
	args := []interface{}{"multi_get"}

	for _, v := range key {
		args = append(args, v)
	}

	resp, err := c.Do(args...)

	if err != nil {
		return nil, nil, goerr.Errorf(err, "MultiGet %s error", key)
	}

	size := len(resp)
	if size > 0 && resp[0] == oK {

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

//MultiGetArray 批量获取一批 key 对应的值内容.（输入分片）,MultiGet的别名
//
//  key，要获取的 key，可以为多个
//  返回 val，一个包含返回的 map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiGetArray(key []string) (val map[string]Value, err error) {
	return c.MultiGet(key...)
}

//MultiGetSliceArray 批量获取一批 key 对应的值内容.（输入分片）,MultiGetSlice的别名
//
//  key，要获取的 key，可以为多个
//  返回 keys和value分片
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiGetSliceArray(key []string) (keys []string, values []Value, err error) {
	return c.MultiGetSlice(key...)
}

//MultiDel 批量删除一批 key 和其对应的值内容.
//
//  key，要删除的 key，可以为多个
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiDel(key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	args := []interface{}{"multi_del"}
	for _, v := range key {
		args = append(args, v)
	}
	resp, err := c.Do(args...)
	if err != nil {
		return goerr.Errorf(err, "MultiDel %s error", key)
	}

	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, key)
}

//Setbit 设置字符串内指定位置的位值(BIT), 字符串的长度会自动扩展.
//
//  key 键值
//  offset 位偏移
//  bit  0 或 1
//  返回 val，原来的位值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Setbit(key string, offset int64, bit int) (uint, error) {

	resp, err := c.Do("setbit", key, offset, bit)

	if err != nil {
		return 0, goerr.Errorf(err, "Setbit %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]).UInt(), nil
	}
	return 0, makeError(resp, key)
}

//Getbit 获取字符串内指定位置的位值(BIT).
//
//  key 键值
//  offset 位偏移
//  返回 val，位值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Getbit(key string, offset int64) (uint, error) {

	resp, err := c.Do("getbit", key, offset)

	if err != nil {
		return 0, goerr.Errorf(err, "Getbit %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]).UInt(), nil
	}
	return 0, makeError(resp, key)
}

//BitCount 计算字符串的子串所包含的位值为 1 的个数. 若 start 是负数, 则从字符串末尾算起. 若 end 是负数, 则表示从字符串末尾算起(包含). 类似 Redis 的 bitcount
//
//  key 键值
//  start 子串的字节偏移
//  end 子串的字节结尾
//  返回 val，位值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) BitCount(key string, start int64, end int64) (uint, error) {
	resp, err := c.Do("bitcount", key, start, end)
	if err != nil {
		return 0, goerr.Errorf(err, "BitCount %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		//fmt.Println(Value(resp[1]).String())
		return Value(resp[1]).UInt(), nil
	}
	return 0, makeError(resp, key)
}

//CountBit 计算字符串的子串所包含的位值为 1 的个数. 若 start 是负数, 则从字符串末尾算起. 若 size 是负数, 则表示从字符串末尾算起, 忽略掉那么多字节.
//
//  key 键值
//  start 子串的字节偏移
//  size 子串的字节结尾
//  返回 val，位值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) CountBit(key string, start int64, size int64) (byte, error) {
	resp, err := c.Do("countbit", key, start, size)
	if err != nil {
		return 0, goerr.Errorf(err, "CountBit %s error", key)
	}
	if len(resp) == 2 && resp[0] == oK {
		//fmt.Println(Value(resp[1]).String())
		return Value(resp[1]).Byte(), nil
	}
	return 0, makeError(resp, key)
}

//Substr 获取字符串的子串.
//
//  key 键值
//  start int, 子串的字节偏移;若 start 是负数, 则从字符串末尾算起.
//  size  int,可选, 子串的长度(字节数), 默认为到字符串最后一个字节;若 size 是负数, 则表示从字符串末尾算起, 忽略掉那么多字节(类似 PHP 的 substr())
//  返回 val，字符串的部分
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Substr(key string, start int64, size ...int64) (val string, err error) {
	var resp []string
	if len(size) > 0 {
		resp, err = c.Do("substr", key, start, size[0])
	} else {
		resp, err = c.Do("substr", key, start)
	}

	if err != nil {
		return "", goerr.Errorf(err, "Substr %s error", key)
	}
	if len(resp) > 1 && resp[0] == oK {
		return resp[1], nil
	}
	return "", makeError(resp, key)
}

//StrLen 计算字符串的长度(字节数).
//
//  key 键值
//  返回 字符串的长度, key 不存在则返回 0.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) StrLen(key string) (int64, error) {

	resp, err := c.Do("strlen", key)

	if err != nil {
		return -1, goerr.Errorf(err, "Strlen %s error", key)
	}
	if len(resp) > 1 && resp[0] == oK {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, key)
}

//Keys 列出处于区间 (key_start, key_end] 的 key 列表.("", ""] 表示整个区间.
//
//  keyStart int 返回的起始 key(不包含), 空字符串表示 -inf.
//  keyEnd int 返回的结束 key(包含), 空字符串表示 +inf.
//  limit int 最多返回这么多个元素.
//  返回 返回包含 key 的数组.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Keys(keyStart, keyEnd string, limit int64) ([]string, error) {

	resp, err := c.Do("keys", keyStart, keyEnd, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Keys %s %s error", keyStart, keyEnd)
	}
	if len(resp) > 0 && resp[0] == oK {
		return resp[1:], nil
	}
	return nil, makeError(resp, keyStart, keyEnd, limit)
}

//RKeys 列出处于区间 (key_start, key_end] 的 key 列表.("", ""] 表示整个区间.反向选择
//
//  keyStart int 返回的起始 key(不包含), 空字符串表示 -inf.
//  keyEnd int 返回的结束 key(包含), 空字符串表示 +inf.
//  limit int 最多返回这么多个元素.
//  返回 返回包含 key 的数组.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) RKeys(keyStart, keyEnd string, limit int64) ([]string, error) {

	resp, err := c.Do("rkeys", keyStart, keyEnd, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Rkeys %s %s error", keyStart, keyEnd)
	}
	if len(resp) > 0 && resp[0] == oK {
		return resp[1:], nil
	}
	return nil, makeError(resp, keyStart, keyEnd, limit)
}

//Scan 列出处于区间 (key_start, key_end] 的 key-value 列表.("", ""] 表示整个区间.
//
//  keyStart int 返回的起始 key(不包含), 空字符串表示 -inf.
//  keyEnd int 返回的结束 key(包含), 空字符串表示 +inf.
//  limit int 最多返回这么多个元素.
//  返回 返回包含 key 的数组.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) Scan(keyStart, keyEnd string, limit int64) (map[string]Value, error) {

	resp, err := c.Do("scan", keyStart, keyEnd, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Scan %s %s error", keyStart, keyEnd)
	}
	if len(resp) > 0 && resp[0] == oK {
		re := make(map[string]Value)
		size := len(resp)
		for i := 1; i < size-1; i += 2 {
			re[resp[i]] = Value(resp[i+1])
		}
		return re, nil
	}
	return nil, makeError(resp, keyStart, keyEnd, limit)
}

//RScan 列出处于区间 (key_start, key_end] 的 key-value 列表, 反向顺序.("", ""] 表示整个区间.
//
//  keyStart int 返回的起始 key(不包含), 空字符串表示 -inf.
//  keyEnd int 返回的结束 key(包含), 空字符串表示 +inf.
//  limit int 最多返回这么多个元素.
//  返回 返回包含 key 的数组.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) RScan(keyStart, keyEnd string, limit int64) (map[string]Value, error) {

	resp, err := c.Do("rscan", keyStart, keyEnd, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Rscan %s %s error", keyStart, keyEnd)
	}
	if len(resp) > 0 && resp[0] == oK {
		re := make(map[string]Value)
		size := len(resp)
		for i := 1; i < size-1; i += 2 {
			re[resp[i]] = Value(resp[i+1])
		}
		return re, nil
	}
	return nil, makeError(resp, keyStart, keyEnd, limit)
}
