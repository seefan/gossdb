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

//返回 zset 中的元素个数.
//
//  name zset的名称.
//  返回 val 返回包含名字元素的个数.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zsize(name string) (val int64, err error) {
	resp, err := this.Do("zsize", name)
	if err != nil {
		return 0, goerr.NewError(err, "Zsize %s  error", name)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		val = to.Int64(resp[1])
		return val, nil
	}
	return 0, makeError(resp, name)
}

//列出 zset 中的 key 列表. 参见 zscan().
//
//  setName zset名称
//  keyStart score_start 对应的 key.
//  scoreStart 返回 key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd 返回 key 的最大权重值(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 keys 返回符合条件的 key 的数组.
//  返回 scores 返回符合条件的 key 对应的权重.
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zkeys(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (keys []string, err error) {
	resp, err := this.Do("zkeys", setName, keyStart, this.encoding(scoreStart, false), this.encoding(scoreEnd, false), limit)

	if err != nil {
		return nil, goerr.NewError(err, "Zkeys %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		size := len(resp)
		keys := []string{}

		for i := 1; i < size; i++ {
			keys = append(keys, resp[i])
		}
		return keys, nil
	}
	return nil, makeError(resp, setName, keyStart, scoreStart, scoreEnd, limit)
}

//返回 key 处于区间 [start,end] 的 score 的和.
//
//  setName zset名称
//  scoreStart  key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd  key 的最大权重值(包含), 空字符串表示 +inf.
//  返回 val 符合条件的 score 的求和
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zsum(setName string, scoreStart, scoreEnd interface{}) (val int64, err error) {
	resp, err := this.Do("zsum", setName, this.encoding(scoreStart, false), this.encoding(scoreEnd, false))

	if err != nil {
		return 0, goerr.NewError(err, "Zsum %s %v %v  error", setName, scoreStart, scoreEnd)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		val = to.Int64(resp[1])
		return val, nil
	}
	return 0, makeError(resp, setName, scoreStart, scoreEnd)
}

//返回 key 处于区间 [start,end] 的 score 的平均值.
//
//  setName zset名称
//  scoreStart  key 的最小权重值(可能不包含, 依赖 key_start), 空字符串表示 -inf.
//  scoreEnd  key 的最大权重值(包含), 空字符串表示 +inf.
//  返回 val 符合条件的 score 的平均值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zavg(setName string, scoreStart, scoreEnd interface{}) (val int64, err error) {
	resp, err := this.Do("zavg", setName, this.encoding(scoreStart, false), this.encoding(scoreEnd, false))

	if err != nil {
		return 0, goerr.NewError(err, "Zavg %s %v %v  error", setName, scoreStart, scoreEnd)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		val = to.Int64(resp[1])
		return val, nil
	}
	return 0, makeError(resp, setName, scoreStart, scoreEnd)
}

//返回指定 key 在 zset 中的排序位置(排名), 排名从 0 开始. 注意! 本方法可能会非常慢! 请在离线环境中使用.
//
//  setName zset名称
//  key 指定key名
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zrank(setName, key string) (val int64, err error) {
	resp, err := this.Do("zrank", setName, key)

	if err != nil {
		return 0, goerr.NewError(err, "Zrank %s %s  error", setName, key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		val = to.Int64(resp[1])
		return val, nil
	}
	return 0, makeError(resp, setName, key)
}

//返回指定 key 在 zset 中的倒序排名.注意! 本方法可能会非常慢! 请在离线环境中使用.
//
//  setName zset名称
//  key 指定key名
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zrrank(setName, key string) (val int64, err error) {
	resp, err := this.Do("zrrank", setName, key)

	if err != nil {
		return 0, goerr.NewError(err, "Zrrank %s %s  error", setName, key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		val = to.Int64(resp[1])
		return val, nil
	}
	return 0, makeError(resp, setName, key)
}

//根据下标索引区间 [offset, offset + limit) 获取 key-score 对, 下标从 0 开始.注意! 本方法在 offset 越来越大时, 会越慢!
//
//  setName zset名称
//  offset 从此下标处开始返回. 从 0 开始.
//  limit  最多返回这么多个 key-score 对.
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zrange(setName string, offset, limit int64) (val map[string]int64, err error) {
	resp, err := this.Do("zrange", setName, this.encoding(offset), this.encoding(limit))

	if err != nil {
		return nil, goerr.NewError(err, "Zrange %s %s  error", setName, offset, limit)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		val = make(map[string]int64)
		size := len(resp)

		for i := 1; i < size-1; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, offset, limit)
}

//根据下标索引区间 [offset, offset + limit) 获取 key-score 对, 反向顺序获取.注意! 本方法在 offset 越来越大时, 会越慢!
//
//  setName zset名称
//  offset 从此下标处开始返回. 从 0 开始.
//  limit  最多返回这么多个 key-score 对.
//  返回 val 排名
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zrrange(setName string, offset, limit int64) (val map[string]int64, err error) {
	resp, err := this.Do("zrrange", setName, this.encoding(offset), this.encoding(limit))

	if err != nil {
		return nil, goerr.NewError(err, "Zrrange %s %s  error", setName, offset, limit)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		val = make(map[string]int64)
		size := len(resp)

		for i := 1; i < size-1; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, offset, limit)
}

//删除位置处于区间 [start,end] 的元素.
//
//  setName zset名称
//  start 区间开始，包含start值
//  end  区间结束，包含end值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zremrangebyrank(setName string, start, end int64) (err error) {
	resp, err := this.Do("zremrangebyrank", setName, this.encoding(start), this.encoding(end))

	if err != nil {
		return goerr.NewError(err, "Zremrangebyrank %s %s  error", setName, start, end)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, start, end)
}

//删除权重处于区间 [start,end] 的元素.
//
//  setName zset名称
//  start 区间开始，包含start值
//  end  区间结束，包含end值
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zremrangebyscore(setName string, start, end int64) (err error) {
	resp, err := this.Do("zremrangebyscore", setName, this.encoding(start), this.encoding(end))

	if err != nil {
		return goerr.NewError(err, "Zremrangebyscore %s %s  error", setName, start, end)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, start, end)
}

//从 zset 首部删除并返回 `limit` 个元素.
//
//  setName zset名称
//  limit 最多要删除并返回这么多个 key-score 对.
//  返回 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zpopfront(setName string, limit int64) (val map[string]int64, err error) {
	resp, err := this.Do("zpop_front", setName, this.encoding(limit))

	if err != nil {
		return nil, goerr.NewError(err, "Zpopfront %s %s  error", setName, limit)
	}
	size := len(resp)
	if size > 0 && resp[0] == "ok" {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, limit)
}

//从 zset 尾部删除并返回 `limit` 个元素.
//
//  setName zset名称
//  limit 最多要删除并返回这么多个 key-score 对.
//  返回 包含 key-score 的map
//  返回 err，可能的错误，操作成功返回 nil
func (this *Client) Zpopback(setName string, limit int64) (val map[string]int64, err error) {
	resp, err := this.Do("zpop_back", setName, this.encoding(limit))

	if err != nil {
		return nil, goerr.NewError(err, "Zpopback %s %s  error", setName, limit)
	}
	size := len(resp)
	if size > 0 && resp[0] == "ok" {
		val = make(map[string]int64)
		for i := 1; i < size && i+1 < size; i += 2 {
			val[resp[i]] = Value(resp[i+1]).Int64()
		}
		return val, nil
	}
	return nil, makeError(resp, setName, limit)
}
