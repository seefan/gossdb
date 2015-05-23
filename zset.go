package gossdb

import (
	"github.com/seefan/goerr"
	"log"
)

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

// scoreStart,scoreEnd 空字符串"" 或者 int64
func (this *Client) Zscan(setName string, keyStart string, scoreStart, scoreEnd interface{}, limit int64) (map[string]int64, error) {
	resp, err := this.Client.Do("zscan", setName, keyStart, this.encoding(scoreStart, false), this.encoding(scoreEnd, false), limit)

	if err != nil {
		return nil, goerr.NewError(err, "Zscan %s %v %v %v %v error", setName, keyStart, scoreStart, scoreEnd, limit)
	}
	if len(resp) > 0 && resp[0] == "ok" {
		re := make(map[string]int64)
		size := len(resp)
		for i := 1; i < size-1; i += 2 {
			re[resp[i]] = Value(resp[i+1]).Int64()
		}
		return re, nil
	}
	return nil, makeError(resp, setName, keyStart, scoreStart, scoreEnd, limit)
}

func (this *Client) MultiZset(setName string, kvs map[string]int64) (err error) {

	args := []string{}
	for k, v := range kvs {
		args = append(args, k)
		args = append(args, this.encoding(v, false))
	}
	resp, err := this.Client.Do("multi_zset", setName, args)

	if err != nil {
		return goerr.NewError(err, "MultiZset %s %s error", setName, kvs)
	}

	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, setName, kvs)
}

func (this *Client) MultiZget(setName string, key ...string) (val map[string]int64, err error) {
	if len(key) == 0 {
		return make(map[string]int64), nil
	}
	resp, err := this.Client.Do("multi_zget", setName, key)

	if err != nil {
		return nil, goerr.NewError(err, "MultiZget %s %s error", setName, key)
	}
	log.Println("MultiZget", resp)
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

func (this *Client) MultiZdel(setName string, key ...string) (err error) {
	if len(key) == 0 {
		return nil
	}
	resp, err := this.Client.Do("multi_zdel", key)

	if err != nil {
		return goerr.NewError(err, "MultiZdel %s %s error", setName, key)
	}
	log.Println("MultiZdel", resp)
	if len(resp) > 0 && resp[0] == "ok" {
		return nil
	}
	return makeError(resp, key)
}
