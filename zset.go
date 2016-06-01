package gossdb

import (
	"github.com/seefan/goerr"
	"github.com/seefan/to"
)

//设置 zset 中指定 key 对应的权重值.
//
//  setName zset名称
//  key zset 中的 key.
//  score 整数, key 对应的权重值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zset(setName, key string, score int64) (err error) {
	resp, err := this.Do("zset", setName, key, this.encoding(score, false))
	if err != nil {
		return goerr.NewError(err, "Zset %s/%s error", setName, key)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, key)
}

//获取 zset 中指定 key 对应的权重值.
//
//  setName zset名称
//  key zset 中的 key.
//  返回 score 整数, key 对应的权重值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zget(setName, key string) (score int64, err error) {
	resp, err := this.Do("zget", setName, key)
	if err != nil {
		return 0, goerr.NewError(err, "Zget %s/%s error", setName, key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return 0, makeError(resp, setName, key)
}

//删除 zset 中指定 key
//
//  setName zset名称
//  key zset 中的 key.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zdel(setName, key string) (err error) {
	resp, err := this.Do("zdel", setName, key)
	if err != nil {
		return goerr.NewError(err, "Zdel %s/%s error", setName, key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, key)
}

//判断指定的 key 是否存在于 zset 中.
//
//  setName zset名称
//  key zset 中的 key.
//  返回 re 如果存在, 返回 true, 否则返回 false.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zexists(setName, key string) (re bool, err error) {
	resp, err := this.Do("zexists", setName, key)
	if err != nil {
		return false, goerr.NewError(err, "Zexists %s/%s error", setName, key)
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return resp[1] == "1", nil
	}
	return false, makeError(resp, setName, key)
}

//返回处于区间 [start,end] key 数量.
//
//  setName zset名称
//  start key 的最小权重值(包含), 空字符串表示 -inf.
//  end key 的最大权重值(包含), 空字符串表示 +inf.
//  返回 count 返回符合条件的 key 的数量.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zcount(setName string, start, end interface{}) (count int64, err error) {
	resp, err := this.Do("zcount", setName, this.encoding(start, false), this.encoding(end, false))
	if err != nil {
		return -1, goerr.NewError(err, "Zcount %s %v %v error", setName, start, end)
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, setName)
}

//删除 zset 中的所有 key.
//
//  setName zset名称
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zclear(setName string) (err error) {
	resp, err := this.Do("zclear", setName)
	if err != nil {
		return goerr.NewError(err, "Zclear %s error", setName)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName)
}

//列出 zset 中处于区间 (key_start+score_start, score_end] 的 key-score 列表.
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
func (this *Client) Zscan(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (keys []string, scores []int64, err error) {
	resp, err := this.Do("zscan", setName, keyStart, this.encoding(scoreStart, false), this.encoding(scoreEnd, false), limit)

	if err != nil {
		return nil, nil, goerr.NewError(err, "Zscan %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}
	if len(resp) > 0 && resp[0] == "ok" {
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

//列出 zset 中的 key-score 列表, 反向顺序
//
//  setName zset名称
//  keyStart score_start 对应的 key.
//  scoreStart 返回 key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd 返回 key 的最大权重值(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 keys 返回符合条件的 key 的数组.
//  返回 scores 返回符合条件的 key 对应的权重.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zrscan(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (keys []string, scores []int64, err error) {
	resp, err := this.Do("zrscan", setName, keyStart, this.encoding(scoreStart, false), this.encoding(scoreEnd, false), limit)

	if err != nil {
		return nil, nil, goerr.NewError(err, "Zrscan %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}

	if len(resp) > 0 && resp[0] == "ok" {
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

//批量设置 zset 中的 key-score.
//
//  setName zset名称
//  kvs 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiZset(setName string, kvs map[string]int64) (err error) {

	args := []string{}
	for k, v := range kvs {
		args = append(args, k)
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Do("multi_zset", setName, args)

	if err != nil {
		return goerr.NewError(err, "MultiZset %s %s error", setName, kvs)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, kvs)
}

//批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的列表，支持多个key
//  返回 val 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiZget(setName string, key ...string) (val map[string]int64, err error) {
	if len(key) == 0 {
		return make(map[string]int64), nil
	}
	resp, err := this.Do("multi_zget", setName, key)

	if err != nil {
		return nil, goerr.NewError(err, "MultiZget %s %s error", setName, key)
	}
	size := len(resp)
	if size > 0 && resp[0] == "ok" {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, key)
}

//批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的列表，支持多个key
//  返回 keys 包含 key的slice
//  返回 scores 包含 key对应权重的slice
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiZgetSlice(setName string, key ...string) (keys []string, scores []int64, err error) {
	if len(key) == 0 {
		return []string{}, []int64{}, nil
	}
	resp, err := this.Do("multi_zget", setName, key)

	if err != nil {
		return nil, nil, goerr.NewError(err, "MultiZget %s %s error", setName, key)
	}

	size := len(resp)
	if size > 0 && resp[0] == "ok" {

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

//批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的slice
//  返回 val 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiZgetArray(setName string, key []string) (val map[string]int64, err error) {
	if len(key) == 0 {
		return make(map[string]int64), nil
	}
	resp, err := this.Do("multi_zget", setName, key)

	if err != nil {
		return nil, goerr.NewError(err, "MultiZget %s %s error", setName, key)
	}
	size := len(resp)
	if size > 0 && resp[0] == "ok" {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, key)
}

//批量获取 zset 中的 key-score.
//
//  setName zset名称
//  key 要获取key的slice
//  返回 keys 包含 key的slice
//  返回 scores 包含 key对应权重的slice
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiZgetSliceArray(setName string, key []string) (keys []string, scores []int64, err error) {
	if len(key) == 0 {
		return []string{}, []int64{}, nil
	}
	resp, err := this.Do("multi_zget", setName, key)

	if err != nil {
		return nil, nil, goerr.NewError(err, "MultiZget %s %s error", setName, key)
	}

	size := len(resp)
	if size > 0 && resp[0] == "ok" {

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

//批量删除 zset 中的 key-score.
//
//  setName zset名称
//  key 要删除key的列表，支持多个key
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) MultiZdel(setName string, key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	resp, err := this.Do("multi_zdel", key)

	if err != nil {
		return goerr.NewError(err, "MultiZdel %s %s error", setName, key)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, key)
}

//使 zset 中的 key 对应的值增加 num. 参数 num 可以为负数.
//
//  setName zset名称
//  key 要增加权重的key
//  num 要增加权重值
//  返回 int64 增加后的新权重值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zincr(setName string, key string, num int64) (int64, error) {
	if len(key) == 0 {
		return 0, nil
	}
	resp, err := this.Do("zincr", setName, key, this.encoding(num, false))
	if err != nil {
		return 0, goerr.NewError(err, "Zincr %s %s %v", setName, key, num)
	}

	if len(resp) > 1 && resp[0] == "ok" {
		return to.Int64(resp[1]), nil
	}
	return 0, makeError(resp, setName, key)
}

//列出名字处于区间 (name_start, name_end] 的 zset.
//
//  name_start - 返回的起始名字(不包含), 空字符串表示 -inf.
//  name_end - 返回的结束名字(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 []string 返回包含名字的slice.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zlist(nameStart, nameEnd string, limit int64) ([]string, error) {
	resp, err := this.Do("zlist", nameStart, nameEnd, this.encoding(limit, false))
	if err != nil {
		return nil, goerr.NewError(err, "Zlist %s %s %v error", nameStart, nameEnd, limit)
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
