// Package ssdbclient Establish a connection with SSDB, parse the data and convert it into a regular format
//
// 与ssdb建立连接，对数据进行解析，转换成常规格式
package ssdbclient

import (
	"fmt"

	"github.com/seefan/goerr"
	"github.com/seefan/gossdb/v2/conf"
)

const (
	endN = '\n'
	endR = '\r' //在某些版本上会出现
	ok   = "ok"
)

// NewSSDBClient create new ssdb client
//
//	@param cfg initial parameters
//	@return new *SSDBClient
//
// 使用配置创建一个新的SSDBClient，并不实际打开连接
func NewSSDBClient(cfg *conf.Config) *SSDBClient {
	return &SSDBClient{
		connection: connection{
			host:            cfg.Host,
			port:            cfg.Port,
			readTimeout:     cfg.ReadTimeout,
			writeTimeout:    cfg.WriteTimeout,
			readBufferSize:  cfg.ReadBufferSize,
			writeBufferSize: cfg.WriteBufferSize,
			connectTimeout:  cfg.ConnectTimeout,
		},
		retryEnabled: cfg.RetryEnabled,
		password:     cfg.Password,
		encoding:     cfg.Encoding,
	}
}

// SSDBClient ssdb client
type SSDBClient struct {
	//内嵌连接
	connection
	//是否重试
	retryEnabled bool
	//是否自动转码
	encoding bool
	//whether the connection is open
	isOpen   bool
	password string
	//The input parameter is converted to [] bytes, which by default is converted to json format
	//and can be modified to use a custom serialization
	//将输入参数成[]byte，默认会转换成json格式,可以修改这个参数以便使用自定义的序列化方式
	EncodingFunc func(v interface{}) []byte
}

// Start start socket
//
//	@return error that may occur on startup. Return nil if successful startup
//
// 启动连接，并设置读写的缓存
func (s *SSDBClient) Start() error {
	if err := s.start(); err != nil {
		return err
	}
	if s.encoding {
		s.connection.encodingFunc = func(v interface{}) []byte {
			if s.EncodingFunc != nil {
				return s.EncodingFunc(v)
			}
			return nil
		}
	}
	s.isOpen = true
	return s.auth()
}

// Close close SSDBClient
//
//	@return error that may occur on shutdown. Return nil if successful shutdown
func (s *SSDBClient) Close() error {
	s.isOpen = false
	return s.close()
}

// IsOpen check if the connection is open
//
//	@return bool returns true if the connection is open
//
// 是否为打开状态
func (s *SSDBClient) IsOpen() bool {
	return s.isOpen
}

// 执行ssdb命令
func (s *SSDBClient) do(args ...interface{}) (resp []string, err error) {
	if !s.isOpen {
		return nil, goerr.String("gossdb client is closed.")
	}
	defer func() {
		if e := recover(); e != nil {
			s.isOpen = false
			err = fmt.Errorf("%v", e)
		}
	}()
	if err = s.send(args); err != nil {
		s.isOpen = false
		return nil, goerr.Errorf(err, "client send error")
	}
	if resp, err = s.recv(); err != nil {
		s.isOpen = false
		return nil, goerr.Errorf(err, "client recv error")
	}
	return
}
func (s *SSDBClient) auth() error {
	if s.password == "" { //without a password, authentication is not required
		return nil
	}
	//if !s.isAuth {
	resp, err := s.do("auth", s.password)
	if err != nil {
		if e := s.Close(); e != nil {
			err = goerr.Errorf(err, "client close failed")
		}
		return goerr.Errorf(err, "authentication failed")
	}
	if len(resp) > 0 && resp[0] == ok {
		//验证成功
		//s.isAuth = true
		return nil
	}
	return goerr.String("authentication failed,password is wrong")

	//}
	//return nil
}

// Do common function
//
//	@param args the input parameters
//	@return []string output parameters
//	@return error Possible errors
//
// 通用调用方法，所有操作ssdb的函数最终都是调用这个函数
func (s *SSDBClient) Do(args ...interface{}) ([]string, error) {
	//if err := s.auth(); err != nil {
	//	return nil, err
	//}
	resp, err := s.do(args...)
	if err != nil {
		if e := s.Close(); e != nil {
			err = goerr.Errorf(err, "client close failed")
		}
		if s.retryEnabled { //如果允许重试，就重新打开一次连接
			if err = s.Start(); err == nil {
				resp, err = s.do(args...)
				if err != nil {
					if e := s.Close(); e != nil {
						err = goerr.Errorf(err, "client close failed")
					}
				}
			}
		}
	}
	return resp, err
}
