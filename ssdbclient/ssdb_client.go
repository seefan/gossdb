//Establish a connection with SSDB, parse the data and convert it into a regular format
//
//与ssdb建立连接，对数据进行解析，转换成常规格式
package ssdbclient

import (
	"bytes"
	"errors"
	//"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/seefan/goerr"
	"github.com/seefan/gossdb/conf"
	//"github.com/seefan/gossdb"
)

const (
	endN = '\n'
	endR = '\r'
	oK   = "ok"
)

//NewSSDBClient create new ssdb client
//
//  @param cfg initial parameters
//  @return new *SSDBClient
//
//使用配置创建一个新的SSDBClient，并不实际打开连接
func NewSSDBClient(cfg *conf.Config) *SSDBClient {
	return &SSDBClient{
		host:            cfg.Host,
		port:            cfg.Port,
		password:        cfg.Password,
		readTimeout:     cfg.ReadTimeout,
		writeTimeout:    cfg.WriteTimeout,
		readBufferSize:  cfg.ReadBufferSize,
		writeBufferSize: cfg.WriteBufferSize,
		retryEnabled:    cfg.RetryEnabled,
		connectTimeout:  cfg.ConnectTimeout,
		encoding:        cfg.Encoding,
	}
}

//SSDBClient
type SSDBClient struct {
	isOpen   bool
	password string
	host     string
	port     int

	sock      *net.TCPConn
	readBuf   []byte
	packetBuf bytes.Buffer
	//packetBuf bytes.Buffer
	//连接写缓冲，默认为8k，单位为kb
	writeBufferSize int
	//连接读缓冲，默认为8k，单位为kb
	readBufferSize int
	//是否重试
	retryEnabled bool
	//写超时
	writeTimeout int
	//读超时
	readTimeout int
	//0时间
	timeZero time.Time
	//创建连接的超时时间，单位为秒。默认值: 5
	connectTimeout int
	//是否自动转码
	encoding bool
	//The input parameter is converted to [] bytes, which by default is converted to json format
	//and can be modified to use a custom serialization
	//将输入参数成[]byte，默认会转换成json格式,可以修改这个参数以便使用自定义的序列化方式
	EncodingFunc func(v interface{}) []byte
}

//Start start socket
//
//  @return error that may occur on startup. Return nil if successful startup
//
//启动连接，并设置读写的缓存
func (s *SSDBClient) Start() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.host, s.port), time.Second)
	if err != nil {
		return err
	}
	sock := conn.(*net.TCPConn)
	err = sock.SetReadBuffer(s.readBufferSize * 1024)
	if err != nil {
		return err
	}
	err = sock.SetWriteBuffer(s.writeBufferSize * 1024)
	if err != nil {
		return err
	}
	s.readBuf = make([]byte, s.readBufferSize*1024)
	s.sock = sock
	s.timeZero = time.Time{}
	s.isOpen = true
	return nil
}

//Close close SSDBClient
//
//  @return error that may occur on shutdown. Return nil if successful shutdown
func (s *SSDBClient) Close() error {
	s.isOpen = false
	s.readBuf = nil
	if s.sock == nil {
		return nil
	}

	return s.sock.Close()
}

//IsOpen check if the connection is open
//
//  @return bool returns true if the connection is open
//
//是否为打开状态
func (s *SSDBClient) IsOpen() bool {
	return s.isOpen
}

//Ping ping
//
//  @return bool returns true if the connection is normal
//
// 状态检查，正常返回true
func (s *SSDBClient) Ping() bool {
	_, err := s.Do("version")
	return err == nil
}

//执行ssdb命令
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

//Do common function
//
//  @param args the input parameters
//  @return []string output parameters
//  @return error Possible errors
//
//通用调用方法，所有操作ssdb的函数最终都是调用这个函数
func (s *SSDBClient) Do(args ...interface{}) ([]string, error) {
	if s.password != "" {
		resp, err := s.do("auth", []string{s.password})
		if err != nil {
			if e := s.Close(); e != nil {
				err = goerr.Errorf(err, "client close failed")
			}
			return nil, goerr.Errorf(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == oK {
			//验证成功
			s.password = ""
		} else {
			return nil, goerr.String("authentication failed,password is wrong")
		}
	}
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

//write write to buf
func (s *SSDBClient) write(bs []byte) {
	s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
	s.packetBuf.WriteByte(endN)
	s.packetBuf.Write(bs)
	s.packetBuf.WriteByte(endN)
}

//send cmd to ssdb
func (s *SSDBClient) send(args []interface{}) error {
	s.packetBuf.Reset()
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			s.write([]byte(arg))
		case []string:
			key := args[0].(string)
			if strings.HasPrefix(key, "multi") {
				for _, a := range arg {
					s.write([]byte(a))
				}
			} else if s.encoding && s.EncodingFunc != nil {
				s.write(s.EncodingFunc(arg))
			} else {
				s.packetBuf.Reset()
				return errors.New("bad arguments type")
			}
		case []byte:
			s.write(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.write(bs)
		case int8:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.write(bs)
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.write(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.write(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			s.write(bs)
		case uint8:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.write(bs)
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.write(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.write(bs)
		case uint64:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.write(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			s.write(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			s.write(bs)
		case bool:
			if arg {
				s.write([]byte{1})
			} else {
				s.write([]byte{1})
			}
			s.packetBuf.WriteByte(endN)
		case time.Time:
			bs := strconv.AppendInt(nil, arg.Unix(), 10)
			s.write(bs)
		case time.Duration:
			bs := strconv.AppendInt(nil, arg.Nanoseconds(), 10)
			s.write(bs)
		case nil:
			s.write([]byte{})
		default:
			if s.encoding && s.EncodingFunc != nil {
				s.write(s.EncodingFunc(arg))
			} else {
				return errors.New("bad arguments type")
			}
		}

	}
	s.packetBuf.WriteByte(endN)
	if err := s.sock.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(s.writeTimeout))); err != nil {
		return err
	}
	for _, err := s.packetBuf.WriteTo(s.sock); s.packetBuf.Len() > 0; {
		if err != nil {
			s.packetBuf.Reset()
			return goerr.Errorf(err, "client socket write error")
		}
	}
	//设置不超时
	if err := s.sock.SetWriteDeadline(s.timeZero); err != nil {
		return err
	}

	return nil
}

//recv receive data
func (s *SSDBClient) recv() (resp []string, err error) {
	bufSize := -1
	s.packetBuf.Reset()
	//设置读取数据超时，
	if err = s.sock.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.readTimeout))); err != nil {
		return nil, err
	}
	//数据包分解，发现长度，找到结尾，循环发现，发现空行，结束
	end := false
	for !end {
		bufSize, err = s.sock.Read(s.readBuf)
		if err != nil {
			return nil, goerr.Errorf(err, "client socket read error")
		}
		if bufSize < 1 {
			continue
		}
		s.packetBuf.Write(s.readBuf[:bufSize])
		n := bytes.IndexByte(s.readBuf, endN)
		if n == -1 {
			continue
		}

		for {
			rs, re, e := s.parse()
			//end
			if re {
				end = true
				break
			}
			//err
			if e != nil && re {
				err = goerr.Errorf(e, "client socket read error")
				end = true
				break
			}
			//no data
			if err != nil && !re {
				break
			}
			resp = append(resp, rs)
		}
	}
	//设置不超时
	if err = s.sock.SetReadDeadline(s.timeZero); err != nil {
		return nil, err
	}
	return resp, nil
}

//解析数据为string
func (s *SSDBClient) parse() (resp string, end bool, err error) {
	ns, err := s.packetBuf.ReadBytes(endN)
	if err == io.EOF {
		return "", true, nil
	}
	if err != nil {
		return "", true, err
	}
	size := len(ns)
	if size == 1 && ns[0] == endN || size == 2 && ns[0] == endR { //空行，说明一个数据包结束
		return "", true, nil
	}
	blockSize := toNum(ns)
	if ns[size-1] == endR {
		size = 2
	} else {
		size = 1
	}
	if s.packetBuf.Len() < blockSize+size {
		return "", false, io.EOF
	}

	ns = s.packetBuf.Next(blockSize)
	s.packetBuf.Next(size)
	return string(ns), false, nil
}
