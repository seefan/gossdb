package client

import (
	"github.com/seefan/goerr"
)

//ZSet 设置 zset 中指定 key 对应的权重值.
//
//  setName zset名称
//  key zset 中的 key.
//  score 整数, key 对应的权重值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZSet(setName, key string, score int64) (err error) {
	resp, err := c.Do("zset", setName, key, score)
	if err != nil {
		return goerr.Errorf(err, "Zset %s/%s error", setName, key)
	}

	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName, key)
}

//ZGet 获取 zset 中指定 key 对应的权重值.
//
//  setName zset名称
//  key zset 中的 key.
//  返回 score 整数, key 对应的权重值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZGet(setName, key string) (score int64, err error) {
	resp, err := c.Do("zget", setName, key)
	if err != nil {
		return 0, goerr.Errorf(err, "Zget %s/%s error", setName, key)
	}
	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]).Int64(), nil
	}
	return 0, makeError(resp, setName, key)
}

//ZDel 删除 zset 中指定 key
//
//  setName zset名称
//  key zset 中的 key.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZDel(setName, key string) (err error) {
	resp, err := c.Do("zdel", setName, key)
	if err != nil {
		return goerr.Errorf(err, "Zdel %s/%s error", setName, key)
	}
	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName, key)
}

//ZExists 判断指定的 key 是否存在于 zset 中.
//
//  setName zset名称
//  key zset 中的 key.
//  返回 re 如果存在, 返回 true, 否则返回 false.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZExists(setName, key string) (re bool, err error) {
	resp, err := c.Do("zexists", setName, key)
	if err != nil {
		return false, goerr.Errorf(err, "Zexists %s/%s error", setName, key)
	}

	if len(resp) == 2 && resp[0] == oK {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, setName, key)
}

//ZCount 返回处于区间 [start,end] key 数量.
//
//  setName zset名称
//  start key 的最小权重值(包含), 空字符串表示 -inf.
//  end key 的最大权重值(包含), 空字符串表示 +inf.
//  返回 count 返回符合条件的 key 的数量.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZCount(setName string, start, end interface{}) (count int64, err error) {
	resp, err := c.Do("zcount", setName, start, end)
	if err != nil {
		return -1, goerr.Errorf(err, "Zcount %s %v %v error", setName, start, end)
	}

	if len(resp) == 2 && resp[0] == oK {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, setName)
}

//ZClear 删除 zset 中的所有 key.
//
//  setName zset名称
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZClear(setName string) (err error) {
	resp, err := c.Do("zclear", setName)
	if err != nil {
		return goerr.Errorf(err, "Zclear %s error", setName)
	}

	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName)
}

//ZScan 列出 zset 中处于区间 (key_start+score_start, score_end] 的 key-score 列表.
//
//  如果 key_start 为空, 那么对应权重值大于或者等于 score_start 的 key 将被返回. 如果 key_start 不为空, 那么对应权重值大于 score_start 的 key, 或者大于 key_start 且对应权重值等于 score_start 的 key 将被返回.
//  也就是说, 返回的 key 在 (key.score == score_start && key > key_start || key.score > score_start), 并且 key.score <= score_end 区间. 先判断 score_start, score_end, 然后判断 key_start.
//
//  setName zset名称
//  keyStart score_start 对应的 key.
//  scoreStart 返回 key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd 返回 key 的最大权重值(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 keys 返回符合条件的 key 的数组.
//  返回 scores 返回符合条件的 key 对应的权重.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZScan(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (keys []string, scores []int64, err error) {
	resp, err := c.Do("zscan", setName, keyStart, scoreStart, scoreEnd, limit)

	if err != nil {
		return nil, nil, goerr.Errorf(err, "Zscan %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}
	if len(resp) > 0 && resp[0] == oK {
		size := len(resp)
		keys := make([]string, 0, (size-1)/2)
		scores := make([]int64, 0, (size-1)/2)

		for i := 1; i < size-1; i += 2 {
			keys = append(keys, resp[i])
			scores = append(scores, Value(resp[i+1]).Int64())
		}
		return keys, scores, nil
	}
	return nil, nil, makeError(resp, setName, keyStart, scoreStart, scoreEnd, limit)
}

//ZRScan 列出 zset 中的 key-score 列表, 反向顺序
//
//  setName zset名称
//  keyStart score_start 对应的 key.
//  scoreStart 返回 key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd 返回 key 的最大权重值(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 keys 返回符合条件的 key 的数组.
//  返回 scores 返回符合条件的 key 对应的权重.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRScan(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (keys []string, scores []int64, err error) {
	resp, err := c.Do("zrscan", setName, keyStart, scoreStart, scoreEnd, limit)

	if err != nil {
		return nil, nil, goerr.Errorf(err, "Zrscan %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}

	if len(resp) > 0 && resp[0] == oK {
		size := len(resp)
		keys := make([]string, 0, (size-1)/2)
		scores := make([]int64, 0, (size-1)/2)

		for i := 1; i < size-1; i += 2 {
			keys = append(keys, resp[i])
			scores = append(scores, Value(resp[i+1]).Int64())
		}
		return keys, scores, nil
	}
	return nil, nil, makeError(resp, setName, keyStart, scoreStart, scoreEnd, limit)
}

//MultiZSet 批量设置 zset 中的 key-score.
//
//  setName zset名称
//  kvs 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiZSet(setName string, kvs map[string]int64) (err error) {

	args := []interface{}{"multi_zset", setName}
	for k, v := range kvs {
		args = append(args, k)
		args = append(args, v)
	}
	resp, err := c.Do(args...)

	if err != nil {
		return goerr.Errorf(err, "MultiZset %s %v error", setName, kvs)
	}

	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName, kvs)
}

//MultiZGet 批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的列表，支持多个key
//  返回 val 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiZGet(setName string, key ...string) (val map[string]int64, err error) {
	if len(key) == 0 {
		return make(map[string]int64), nil
	}
	args := []interface{}{"multi_zget", setName}

	for _, v := range key {
		args = append(args, v)
	}

	resp, err := c.Do(args...)

	if err != nil {
		return nil, goerr.Errorf(err, "MultiZget %s %s error", setName, key)
	}
	size := len(resp)
	if size > 0 && resp[0] == oK {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, key)
}

//MultiZGetSlice 批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的列表，支持多个key
//  返回 keys 包含 key的slice
//  返回 scores 包含 key对应权重的slice
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiZGetSlice(setName string, key ...string) (keys []string, scores []int64, err error) {
	if len(key) == 0 {
		return []string{}, []int64{}, nil
	}
	args := []interface{}{"multi_zget", setName}
	for _, v := range key {
		args = append(args, v)
	}
	resp, err := c.Do(args...)

	if err != nil {
		return nil, nil, goerr.Errorf(err, "MultiZget %s %s error", setName, key)
	}

	size := len(resp)
	if size > 0 && resp[0] == oK {

		keys := make([]string, (size-1)/2)
		scores := make([]int64, (size-1)/2)

		for i := 1; i < size && i+1 < size; i += 2 {
			keys = append(keys, resp[i])
			scores = append(scores, Value(resp[i+1]).Int64())
		}
		return keys, scores, nil
	}
	return nil, nil, makeError(resp, setName, key)
}

//MultiZGetArray 批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的slice
//  返回 val 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiZGetArray(setName string, key []string) (val map[string]int64, err error) {
	return c.MultiZGet(setName, key...)
}

//MultiZgetSliceArray 批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的slice
//  返回 keys 包含 key的slice
//  返回 scores 包含 key对应权重的slice
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiZgetSliceArray(setName string, key []string) (keys []string, scores []int64, err error) {
	return c.MultiZGetSlice(setName, key...)
}

//MultiZDel 批量删除 zset 中的 key-score.
//
//  setName zset名称
//  key 要删除key的列表，支持多个key
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) MultiZDel(setName string, key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	args := []interface{}{"multi_zdel", setName}
	for _, v := range key {
		args = append(args, v)
	}
	resp, err := c.Do(args...)
	if err != nil {
		return goerr.Errorf(err, "MultiZdel %s %s error", setName, key)
	}

	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName, key)
}

//ZIncr 使 zset 中的 key 对应的值增加 num. 参数 num 可以为负数.
//
//  setName zset名称
//  key 要增加权重的key
//  num 要增加权重值
//  返回 int64 增加后的新权重值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZIncr(setName string, key string, num int64) (int64, error) {
	if len(key) == 0 {
		return 0, nil
	}
	resp, err := c.Do("zincr", setName, key, num)
	if err != nil {
		return 0, goerr.Errorf(err, "Zincr %s %s %v", setName, key, num)
	}

	if len(resp) > 1 && resp[0] == oK {
		return Value(resp[1]).Int64(), nil
	}
	return 0, makeError(resp, setName, key)
}

//ZList 列出名字处于区间 (name_start, name_end] 的 zset.
//
//  name_start - 返回的起始名字(不包含), 空字符串表示 -inf.
//  name_end - 返回的结束名字(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 []string 返回包含名字的slice.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZList(nameStart, nameEnd string, limit int64) ([]string, error) {
	resp, err := c.Do("zlist", nameStart, nameEnd, limit)
	if err != nil {
		return nil, goerr.Errorf(err, "Zlist %s %s %v error", nameStart, nameEnd, limit)
	}

	if len(resp) > 0 && resp[0] == oK {
		size := len(resp)
		keyList := make([]string, 0, size-1)

		for i := 1; i < size; i++ {
			keyList = append(keyList, resp[i])
		}
		return keyList, nil
	}
	return nil, makeError(resp, nameStart, nameEnd, limit)
}

//ZSize 返回 zset 中的元素个数.
//
//  name zset的名称.
//  返回 val 返回包含名字元素的个数.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZSize(name string) (val int64, err error) {
	resp, err := c.Do("zsize", name)
	if err != nil {
		return 0, goerr.Errorf(err, "Zsize %s  error", name)
	}

	if len(resp) > 0 && resp[0] == oK {
		val = Value(resp[1]).Int64()
		return val, nil
	}
	return 0, makeError(resp, name)
}

//ZKeys 列出 zset 中的 key 列表. 参见 zscan().
//
//  setName zset名称
//  keyStart score_start 对应的 key.
//  scoreStart 返回 key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd 返回 key 的最大权重值(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 keys 返回符合条件的 key 的数组.
//  返回 scores 返回符合条件的 key 对应的权重.
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZKeys(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (keys []string, err error) {
	resp, err := c.Do("zkeys", setName, keyStart, scoreStart, scoreEnd, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Zkeys %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}
	if len(resp) > 0 && resp[0] == oK {
		size := len(resp)
		keys := []string{}

		for i := 1; i < size; i++ {
			keys = append(keys, resp[i])
		}
		return keys, nil
	}
	return nil, makeError(resp, setName, keyStart, scoreStart, scoreEnd, limit)
}

//ZSum 返回 key 处于区间 [start,end] 的 score 的和.
//
//  setName zset名称
//  scoreStart  key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd  key 的最大权重值(包含), 空字符串表示 +inf.
//  返回 val 符合条件的 score 的求和
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZSum(setName string, scoreStart, scoreEnd interface{}) (val int64, err error) {
	resp, err := c.Do("zsum", setName, scoreStart, scoreEnd)

	if err != nil {
		return 0, goerr.Errorf(err, "Zsum %s %v %v  error", setName, scoreStart, scoreEnd)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = Value(resp[1]).Int64()
		return val, nil
	}
	return 0, makeError(resp, setName, scoreStart, scoreEnd)
}

//ZAvg 返回 key 处于区间 [start,end] 的 score 的平均值.
//
//  setName zset名称
//  scoreStart  key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd  key 的最大权重值(包含), 空字符串表示 +inf.
//  返回 val 符合条件的 score 的平均值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZAvg(setName string, scoreStart, scoreEnd interface{}) (val int64, err error) {
	resp, err := c.Do("zavg", setName, scoreStart, scoreEnd)

	if err != nil {
		return 0, goerr.Errorf(err, "Zavg %s %v %v  error", setName, scoreStart, scoreEnd)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = Value(resp[1]).Int64()
		return val, nil
	}
	return 0, makeError(resp, setName, scoreStart, scoreEnd)
}

//ZRank 返回指定 key 在 zset 中的排序位置(排名), 排名从 0 开始. 注意! 本方法可能会非常慢! 请在离线环境中使用.
//
//  setName zset名称
//  key 指定key名
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRank(setName, key string) (val int64, err error) {
	resp, err := c.Do("zrank", setName, key)

	if err != nil {
		return 0, goerr.Errorf(err, "Zrank %s %s  error", setName, key)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = Value(resp[1]).Int64()
		return val, nil
	}
	return 0, makeError(resp, setName, key)
}

//ZRRank 返回指定 key 在 zset 中的倒序排名.注意! 本方法可能会非常慢! 请在离线环境中使用.
//
//  setName zset名称
//  key 指定key名
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRRank(setName, key string) (val int64, err error) {
	resp, err := c.Do("zrrank", setName, key)

	if err != nil {
		return 0, goerr.Errorf(err, "Zrrank %s %s  error", setName, key)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = Value(resp[1]).Int64()
		return val, nil
	}
	return 0, makeError(resp, setName, key)
}

//ZRange 根据下标索引区间 [offset, offset + limit) 获取 key-score 对, 下标从 0 开始.注意! 本方法在 offset 越来越大时, 会越慢!
//
//  setName zset名称
//  offset 从此下标处开始返回. 从 0 开始.
//  limit  最多返回这么多个 key-score 对.
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRange(setName string, offset, limit int64) (val map[string]int64, err error) {
	resp, err := c.Do("zrange", setName, offset, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Zrange %s %d %d  error", setName, offset, limit)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = make(map[string]int64)
		size := len(resp)
		for i := 1; i < size-1; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, offset, limit)
}

//ZRangeSlice 根据下标索引区间 [offset, offset + limit) 获取 获取 key和score 数组对, 下标从 0 开始.注意! 本方法在 offset 越来越大时, 会越慢!
//
//  setName zset名称
//  offset 从此下标处开始返回. 从 0 开始.
//  limit  最多返回这么多个 key-score 对.
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRangeSlice(setName string, offset, limit int64) (key []string, val []int64, err error) {
	resp, err := c.Do("zrange", setName, offset, limit)

	if err != nil {
		return nil, nil, goerr.Errorf(err, "Zrange_slice %s %d %d  error", setName, offset, limit)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = []int64{}
		key = []string{}
		size := len(resp)
		for i := 1; i < size-1; i += 2 {
			key = append(key, resp[i])
			val = append(val, Value(resp[i+1]).Int64())
		}
		return key, val, nil
	}
	return nil, nil, makeError(resp, setName, offset, limit)
}

//ZRRange 根据下标索引区间 [offset, offset + limit) 获取 key-score 对, 反向顺序获取.注意! 本方法在 offset 越来越大时, 会越慢!
//
//  setName zset名称
//  offset 从此下标处开始返回. 从 0 开始.
//  limit  最多返回这么多个 key-score 对.
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRRange(setName string, offset, limit int64) (val map[string]int64, err error) {
	resp, err := c.Do("zrrange", setName, offset, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Zrrange %s %d %d  error", setName, offset, limit)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = make(map[string]int64)
		size := len(resp)

		for i := 1; i < size-1; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, offset, limit)
}

//ZRRangeSlice 根据下标索引区间 [offset, offset + limit) 获取 key和score 数组对, 反向顺序获取.注意! 本方法在 offset 越来越大时, 会越慢!
//
//  setName zset名称
//  offset 从此下标处开始返回. 从 0 开始.
//  limit  最多返回这么多个 key-score 对.
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRRangeSlice(setName string, offset, limit int64) (key []string, val []int64, err error) {
	resp, err := c.Do("zrrange", setName, offset, limit)

	if err != nil {
		return nil, nil, goerr.Errorf(err, "Zrrange_slice %s %d %d  error", setName, offset, limit)
	}
	if len(resp) > 0 && resp[0] == oK {
		val = []int64{}
		key = []string{}
		size := len(resp)

		for i := 1; i < size-1; i += 2 {
			key = append(key, resp[i])
			val = append(val, Value(resp[i+1]).Int64())
		}
		return key, val, nil
	}
	return nil, nil, makeError(resp, setName, offset, limit)
}

//ZRemRangeByRank 删除位置处于区间 [start,end] 的元素.
//
//  setName zset名称
//  start 区间开始，包含start值
//  end  区间结束，包含end值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRemRangeByRank(setName string, start, end int64) (err error) {
	resp, err := c.Do("zremrangebyrank", setName, start, end)

	if err != nil {
		return goerr.Errorf(err, "Zremrangebyrank %s %d %d  error", setName, start, end)
	}
	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName, start, end)
}

//ZRemRangeByScore 删除权重处于区间 [start,end] 的元素.
//
//  setName zset名称
//  start 区间开始，包含start值
//  end  区间结束，包含end值
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZRemRangeByScore(setName string, start, end int64) (err error) {
	resp, err := c.Do("zremrangebyscore", setName, start, end)

	if err != nil {
		return goerr.Errorf(err, "Zremrangebyscore %s %d %d  error", setName, start, end)
	}
	if len(resp) > 0 && resp[0] == oK {
		return nil
	}
	return makeError(resp, setName, start, end)
}

//ZPopFront 从 zset 首部删除并返回 `limit` 个元素.
//
//  setName zset名称
//  limit 最多要删除并返回这么多个 key-score 对.
//  返回 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZPopFront(setName string, limit int64) (val map[string]int64, err error) {
	resp, err := c.Do("zpop_front", setName, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Zpopfront %s %d   error", setName, limit)
	}
	size := len(resp)
	if size > 0 && resp[0] == oK {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, limit)
}

//ZPopBack 从 zset 尾部删除并返回 `limit` 个元素.
//
//  setName zset名称
//  limit 最多要删除并返回这么多个 key-score 对.
//  返回 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (c *Client) ZPopBack(setName string, limit int64) (val map[string]int64, err error) {
	resp, err := c.Do("zpop_back", setName, limit)

	if err != nil {
		return nil, goerr.Errorf(err, "Zpopback %s %d   error", setName, limit)
	}
	size := len(resp)
	if size > 0 && resp[0] == oK {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, limit)
}
