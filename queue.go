package gossdb

import (
	"github.com/seefan/goerr"
)

var (
	qtrim_cmd  = []string{"qtrim_front", "qtrim_back"}
	qpush_cmd  = []string{"qpush_front", "qpush_back"}
	qpop_cmd   = []string{"qpop_front", "qpop_back"}
	qslice_cmd = []string{"qslice", "qrange"}
)

//返回队列的长度.
//
//  name  队列的名字
//  返回 size，队列的长度；
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qsize(name string) (size int64, err error) {
	resp, err := this.Do("qsize", name)
	if err != nil {
		return -1, goerr.NewError(err, "Qsize %s error", name)
	}

	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, name)
}

//清空一个队列.
//
//  name  队列的名字
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qclear(name string) (err error) {
	resp, err := this.Do("qclear", name)
	if err != nil {
		return goerr.NewError(err, "Qclear %s error", name)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, name)
}

//往队列的首部添加一个或者多个元素
//
//  name  队列的名字
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpush_front(name string, value ...interface{}) (size int64, err error) {
	return this.qpush(name, false, value...)
}

//往队列的首部添加一个或者多个元素
//
//  name  队列的名字
//  reverse 是否反向
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) qpush(name string, reverse bool, value ...interface{}) (size int64, err error) {
	if len(value) == 0 {
		return -1, nil
	}
	index := 0
	if reverse {
		index = 1
	}
	args := []string{name}
	for _, v := range value {
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Do(qpush_cmd[index], args)
	if err != nil {
		return -1, goerr.NewError(err, "%s %s error", qpush_cmd[index], name)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, name)
}

//往队列的尾部添加一个或者多个元素
//
//  name  队列的名字
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpush(name string, value ...interface{}) (size int64, err error) {
	return this.qpush(name, true, value...)
}

//往队列的尾部添加一个或者多个元素
//
//  name  队列的名字
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpush_back(name string, value ...interface{}) (size int64, err error) {
	return this.qpush(name, true, value...)
}

//从队列首部弹出最后一个元素.
//
//  name 队列的名字
//  返回 v，返回一个元素，并在队列中删除 v；队列为空时返回空值
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpop_front(name string) (v Value, err error) {
	return this.Qpop(name)
}

//从队列尾部弹出最后一个元素.
//
//  name 队列的名字
//  返回 v，返回一个元素，并在队列中删除 v；队列为空时返回空值
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpop_back(name string) (v Value, err error) {
	return this.Qpop(name, true)
}

//从队列首部弹出最后一个元素.
//
//  name 队列的名字
//  返回 v，返回一个元素，并在队列中删除 v；队列为空时返回空值
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpop(name string, reverse ...bool) (v Value, err error) {
	index := 1
	if len(reverse) > 0 && !reverse[0] {
		index = 0
	}
	resp, err := this.Do(qpop_cmd[index], name)
	if err != nil {
		return "", goerr.NewError(err, "%s %s error", qpop_cmd[index], name)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, name)
}

//从队列首部弹出最后多个元素.
//
//  name 队列的名字
//  返回 v，返回多个元素，并在队列中弹出多个元素；
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpop_front_array(name string, size int64) (v []Value, err error) {
	return this.QpopArray(name, size, false)
}

//从队列尾部弹出最后多个元素.
//
//  name 队列的名字
//  返回 v，返回多个元素，并在队列中弹出多个元素；
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpop_back_array(name string, size int64) (v []Value, err error) {
	return this.QpopArray(name, size, true)
}

//从队列首部弹出最后多个个元素.
//
//  name 队列的名字
//  size 取出元素的数量
//  reverse 是否反转取
//  返回 v，返回多个元素，并在队列中弹出多个元素；
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) QpopArray(name string, size int64, reverse ...bool) (v []Value, err error) {
	index := 1
	if len(reverse) > 0 && !reverse[0] {
		index = 0
	}
	resp, err := this.Do(qpop_cmd[index], name, size)
	if err != nil {
		return nil, goerr.NewError(err, "%s %s error", qpop_cmd[index], name)
	}

	respsize := len(resp)
	if respsize > 1 && resp[0] == "ok" {
		for i := 1; i < respsize; i++ {
			v = append(v, Value(resp[i]))
		}
		return
	}
	return nil, makeError(resp, name)
}

//返回下标处于区域 [offset, offset + limit] 的元素.
//
//  name queue 的名字.
//  offset 整数, 从此下标处开始返回. 从 0 开始. 可以是负数, 表示从末尾算起.
//  limit 正整数, 最多返回这么多个元素.
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qrange(name string, offset, limit int) (v []Value, err error) {
	return this.slice(name, offset, limit, 1)
}

//返回下标处于区域 [begin, end] 的元素. begin 和 end 可以是负数
//
//  name queue 的名字.
//  begin 正整数, 从此下标处开始返回。从 0 开始。
//  end 整数, 结束下标。可以是负数, 表示返回所有。
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qslice(name string, begin, end int) (v []Value, err error) {
	return this.slice(name, begin, end, 0)
}

//返回下标处于区域 [begin, end] 的元素. begin 和 end 可以是负数
//
//  name queue 的名字.
//  begin 正整数, 从此下标处开始返回。从 0 开始。
//  end 整数, 结束下标。可以是负数, 表示返回所有。
//  [slice，range] 命令
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) slice(name string, args ...int) (v []Value, err error) {
	begin := 0
	end := -1
	index := 0
	if len(args) > 0 {
		begin = args[0]
	}
	if len(args) > 1 {
		end = args[1]
	}

	if len(args) > 2 {
		index = args[2]
	}
	resp, err := this.Do(qslice_cmd[index], name, begin, end)
	if err != nil {
		return nil, goerr.NewError(err, "%s %s error", qslice_cmd[index], name)
	}
	size := len(resp)
	if size >= 1 && resp[0] == "ok" {
		for i := 1; i < size; i++ {
			v = append(v, Value(resp[i]))
		}
		return
	}
	return nil, makeError(resp, name)
}

//从队列头部删除多个元素.
//
//  name queue 的名字.
//  size 最多从队列删除这么多个元素
//  reverse 可选，是否反向执行
//  返回 delSize，返回被删除的元素数量
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qtrim(name string, size int, reverse ...bool) (delSize int64, err error) {
	index := 0
	if len(reverse) > 0 && reverse[0] {
		index = 1
	}
	resp, err := this.Do(qtrim_cmd[index], name, size)
	if err != nil {
		return -1, goerr.NewError(err, "%s %s error", qtrim_cmd[index], name)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, name)
}

//从队列头部删除多个元素.
//
//  name queue 的名字.
//  size 最多从队列删除这么多个元素
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qtrim_front(name string, size int) (delSize int64, err error) {
	return this.Qtrim(name, size)
}

//从队列尾部删除多个元素.
//
//  name queue 的名字.
//  size 最多从队列删除这么多个元素
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qtrim_back(name string, size int) (delSize int64, err error) {
	return this.Qtrim(name, size, true)
}

//列出名字处于区间 (name_start, name_end] 的 queue/list.
//
//  name_start  返回的起始名字(不包含), 空字符串表示 -inf.
//  name_end  返回的结束名字(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qlist(nameStart, nameEnd string, limit int64) ([]string, error) {
	resp, err := this.Do("qlist", nameStart, nameEnd, this.encoding(limit, false))
	if err != nil {
		return nil, goerr.NewError(err, "Qlist %s %s %v error", nameStart, nameEnd, limit)
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

//列出名字处于区间 (name_start, name_end] 的 queue/list.
//
//  name_start  返回的起始名字(不包含), 空字符串表示 -inf.
//  name_end  返回的结束名字(包含), 空字符串表示 +inf.
//  limit  最多返回这么多个元素.
//  返回 v，返回元素的数组，为空时返回 nil
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qrlist(nameStart, nameEnd string, limit int64) ([]string, error) {
	resp, err := this.Do("qrlist", nameStart, nameEnd, this.encoding(limit, false))
	if err != nil {
		return nil, goerr.NewError(err, "Qrlist %s %s %v error", nameStart, nameEnd, limit)
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

//更新位于 index 位置的元素. 如果超过现有的元素范围, 会返回错误.
//
//  key  队列的名字
//  index 指定的位置，可传负数.
//  val  传入的值.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qset(key string, index int64, val interface{}) (err error) {
	var resp []string

	resp, err = this.Do("qset", key, index, this.encoding(val, false))

	if err != nil {
		return goerr.NewError(err, "Qset %s error", key)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}

//返回指定位置的元素. 0 表示第一个元素, 1 是第二个 ... -1 是最后一个.
//
//  key  队列的名字
//  index 指定的位置，可传负数.
//  返回 val，返回的值.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qget(key string, index int64) (Value, error) {
	resp, err := this.Do("qget", key, index)
	if err != nil {
		return "", goerr.NewError(err, "Qget %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//返回队列的第一个元素.
//
//  key  队列的名字
//  返回 val，返回的值.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qfront(key string) (Value, error) {
	resp, err := this.Do("qfront", key)
	if err != nil {
		return "", goerr.NewError(err, "Qfront %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//返回队列的最后一个元素.
//
//  key  队列的名字
//  返回 val，返回的值.
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qback(key string) (Value, error) {
	resp, err := this.Do("qback", key)
	if err != nil {
		return "", goerr.NewError(err, "Qback %s error", key)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]), nil
	}
	return "", makeError(resp, key)
}

//往队列的首部添加一个或者多个元素
//
//  name  队列的名字
//  reverse 是否反向
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) qpush_array(name string, reverse bool, value []interface{}) (size int64, err error) {
	if len(value) == 0 {
		return -1, nil
	}
	index := 0
	if reverse {
		index = 1
	}
	args := []string{name}
	for _, v := range value {
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Do(qpush_cmd[index], args)
	if err != nil {
		return -1, goerr.NewError(err, "%s %s error", qpush_cmd[index], name)
	}
	if len(resp) == 2 && resp[0] == "ok" {
		return Value(resp[1]).Int64(), nil
	}
	return -1, makeError(resp, name)
}

//往队列的尾部添加一个或者多个元素
//
//  name  队列的名字
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpush_array(name string, value []interface{}) (size int64, err error) {
	return this.qpush_array(name, true, value)
}

//往队列的尾部添加一个或者多个元素
//
//  name  队列的名字
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpush_back_array(name string, value []interface{}) (size int64, err error) {
	return this.qpush_array(name, true, value)
}

//往队列的首部添加一个或者多个元素
//
//  name  队列的名字
//  value  存贮的值，可以为多值.
//  返回 size，添加元素之后, 队列的长度
//  返回 err，执行的错误，操作成功返回 nil
func (this *Client) Qpush_front_array(name string, value []interface{}) (size int64, err error) {
	return this.qpush_array(name, false, value)
}
