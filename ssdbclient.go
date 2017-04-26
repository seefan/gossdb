package gossdb

import (
	"github.com/seefan/goerr"
	"github.com/seefan/gopool"
	"github.com/ssdb/gossdb/ssdb"
)

type ISSDBClient interface {
	gopool.Closed
	IsOpen() bool
	Start() error
	Do(args ...interface{}) ([]string, error)
}
type SSDBClient struct {
	conn     *ssdb.Client
	isOpen   bool
	password string
	host     string
	port     int
}

//打开连接
func (s *SSDBClient) Start() error {
	conn, err := ssdb.Connect(s.host, s.port)
	if err != nil {
		return err
	}
	s.isOpen = true
	s.conn = conn
	return nil
}
func (s *SSDBClient) Close() error {
	s.isOpen = false
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
func (s *SSDBClient) IsOpen() bool {
	return s.isOpen
}

//通用调用方法，如果有需要在所有方法前执行的，可以在这里执行
func (s *SSDBClient) Do(args ...interface{}) ([]string, error) {
	if s.password != "" {
		resp, err := s.conn.Do("auth", []string{s.password})
		if err != nil {
			s.conn.Close()
			s.isOpen = false
			return nil, goerr.NewError(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == "ok" {
			//验证成功
			s.password = ""
		} else {
			return nil, makeError(resp, "Authentication failed,password is wrong")
		}
	}
	resp, err := s.conn.Do(args...)
	if err != nil {
		s.conn.Close()
		s.isOpen = false
	}
	return resp, err
}
