package client

import (
	"github.com/seefan/goerr"
	"github.com/ssdb/gossdb/ssdb"
)

type SSDBClient struct {
	conn     *ssdb.Client
	isOpen   bool
	Password string
	Host     string
	Port     int
}

//打开连接
func (s *SSDBClient) Start() error {
	conn, err := ssdb.Connect(s.Host, s.Port)
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
func (s *SSDBClient) Ping() bool {
	_, err := s.Do("info")
	return err == nil
}

//通用调用方法，如果有需要在所有方法前执行的，可以在这里执行
func (s *SSDBClient) Do(args ...interface{}) ([]string, error) {
	if s.Password != "" {
		resp, err := s.conn.Do("auth", []string{s.Password})
		if err != nil {
			s.conn.Close()
			s.isOpen = false
			return nil, goerr.NewError(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == "ok" {
			//验证成功
			s.Password = ""
		} else {
			return nil, goerr.New("Authentication failed,password is wrong")
		}
	}
	resp, err := s.conn.Do(args...)
	if err != nil {
		s.conn.Close()
		s.isOpen = false
	}
	return resp, err
}
