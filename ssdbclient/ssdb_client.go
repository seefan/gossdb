package ssdbclient

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/seefan/goerr"
	"github.com/seefan/gossdb/conf"
	//"github.com/seefan/gossdb"
)

const (
	ENDN = '\n'
	ENDR = '\r'
	OK   = "ok"
)

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
	}
}

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
}

//打开连接
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

//关闭连接
func (s *SSDBClient) Close() error {
	s.isOpen = false
	s.readBuf = nil
	if s.sock == nil {
		return nil
	}

	return s.sock.Close()
}

//是否为打开状态
func (s *SSDBClient) IsOpen() bool {
	return s.isOpen
}

//状态检查
func (s *SSDBClient) Ping() bool {
	_, err := s.Do("info")
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
	} else {
		return
	}
}

//通用调用方法，如果有需要在所有方法前执行的，可以在这里执行
func (s *SSDBClient) Do(args ...interface{}) ([]string, error) {
	if s.password != "" {
		resp, err := s.do("auth", []string{s.password})
		if err != nil {
			if e := s.Close(); e != nil {
				err = goerr.Errorf(err, "client close failed")
			}
			return nil, goerr.Errorf(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == OK {
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

//发送数据
func (s *SSDBClient) send(args []interface{}) error {
	s.packetBuf.Reset()
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.WriteString(arg)
		case []string:
			for _, a := range arg {
				s.packetBuf.Write(strconv.AppendInt(nil, int64(len(a)), 10))
				s.packetBuf.WriteByte(ENDN)
				s.packetBuf.WriteString(a)
				s.packetBuf.WriteByte(ENDN)
			}
			continue
		case []byte:
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case int8:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case uint8:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case uint64:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case bool:
			s.packetBuf.WriteByte(1)
			s.packetBuf.WriteByte(ENDN)
			if arg {
				s.packetBuf.WriteByte(1)
			} else {
				s.packetBuf.WriteByte(0)
			}
		case time.Time:
			bs := strconv.AppendInt(nil, arg.Unix(), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case time.Duration:
			bs := strconv.AppendInt(nil, arg.Nanoseconds(), 10)
			s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.Write(bs)
		case nil:
			s.packetBuf.WriteByte(0)
			s.packetBuf.WriteByte(ENDN)
			s.packetBuf.WriteString("")
		default:
			//if gossdb.Encoding {
			//	if bs, err := json.Marshal(arg); err == nil {
			//		s.packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			//		s.packetBuf.WriteByte(ENDN)
			//		s.packetBuf.Write(bs)
			//	} else {
			//		return goerr.Errorf(err, "bad arguments type,can not json marshal")
			//	}
			//} else {
			//	return goerr.String("bad arguments type")
			//}
		}
		s.packetBuf.WriteByte(ENDN)
	}
	s.packetBuf.WriteByte(ENDN)
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

//接收数据
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
		n := bytes.IndexByte(s.readBuf, ENDN)
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

//解析数据为string的slice
func (s *SSDBClient) parse() (resp string, end bool, err error) {
	ns, err := s.packetBuf.ReadBytes(ENDN)
	if err == io.EOF {
		return "", true, nil
	}
	if err != nil {
		return "", true, err
	}
	size := len(ns)
	if size == 1 && ns[0] == ENDN || size == 2 && ns[0] == ENDR { //空行，说明一个数据包结束
		return "", true, nil
	}
	blockSize := ToNum(ns)
	if ns[size-1] == ENDR {
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
