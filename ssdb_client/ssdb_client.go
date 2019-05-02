package ssdb_client

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/seefan/goerr"
	//"github.com/seefan/gossdb"
)

const (
	ENDN = '\n'
	ENDR = '\r'
	OK   = "ok"
)

type SSDBClient struct {
	isOpen   bool
	Password string
	Host     string
	Port     int

	sock      *net.TCPConn
	readBuf   []byte
	packetBuf bytes.Buffer
	//packetBuf bytes.Buffer
	//连接写缓冲，默认为8k，单位为kb
	WriteBufferSize int
	//连接读缓冲，默认为8k，单位为kb
	ReadBufferSize int
	//是否重试
	RetryEnabled bool
	//读写超时
	ReadWriteTimeout int
	//写超时
	WriteTimeout int
	//读超时
	ReadTimeout int
	//0时间
	timeZero time.Time
	//创建连接的超时时间，单位为秒。默认值: 5
	ConnectTimeout int
}

//打开连接
func (s *SSDBClient) Start() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), time.Second)
	if err != nil {
		return err
	}
	sock := conn.(*net.TCPConn)
	err = sock.SetReadBuffer(s.ReadBufferSize * 1024)
	if err != nil {
		return err
	}
	err = sock.SetWriteBuffer(s.WriteBufferSize * 1024)
	if err != nil {
		return err
	}
	s.readBuf = make([]byte, s.ReadBufferSize*1024)
	s.sock = sock
	s.timeZero = time.Time{}
	s.isOpen = true
	if s.ReadTimeout == 0 {
		s.ReadTimeout = s.ReadWriteTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = s.ReadWriteTimeout
	}
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
func (s *SSDBClient) do(args ...interface{}) ([]string, error) {
	if !s.isOpen {
		return nil, goerr.String("gossdb client is closed.")
	}
	err := s.send(args)
	if err != nil {
		return nil, goerr.Errorf(err, "client send error")
	}
	if resp, err := s.Recv(); err != nil {
		return nil, goerr.Errorf(err, "client recv error")
	} else {
		return resp, nil
	}
}

//通用调用方法，如果有需要在所有方法前执行的，可以在这里执行
func (s *SSDBClient) Do(args ...interface{}) ([]string, error) {
	if s.Password != "" {
		resp, err := s.do("auth", []string{s.Password})
		if err != nil {
			if e := s.Close(); e != nil {
				err = goerr.Errorf(err, "client close failed")
			}
			return nil, goerr.Errorf(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == OK {
			//验证成功
			s.Password = ""
		} else {
			return nil, goerr.String("authentication failed,password is wrong")
		}
	}
	resp, err := s.do(args...)
	if err != nil {
		if e := s.Close(); e != nil {
			err = goerr.Errorf(err, "client close failed")
		}
		if s.RetryEnabled { //如果允许重试，就重新打开一次连接
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
func (s *SSDBClient) Send(args ...interface{}) error {
	if err := s.send(args); err != nil {
		if e := s.Close(); e != nil {
			err = goerr.Errorf(err, "client close failed")
		}
		return err
	}
	return nil
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
	if err := s.sock.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(s.WriteTimeout))); err != nil {
		return err
	}
	for _, err := s.packetBuf.WriteTo(s.sock); s.packetBuf.Len() > 0; {
		if err != nil {
			s.packetBuf.Reset()
			s.isOpen = false
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
func (s *SSDBClient) Recv() (resp []string, err error) {
	bufSize := -1
	s.packetBuf.Reset()
	//设置读取数据超时，
	if err = s.sock.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ReadWriteTimeout))); err != nil {
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
			rs, e, re := s.parse()
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
func (s *SSDBClient) parse() (resp string, err error, end bool) {
	ns, err := s.packetBuf.ReadBytes(ENDN)
	if err == io.EOF {
		return "", nil, true
	}
	if err != nil {
		return "", err, true
	}
	size := len(ns)
	if size == 1 && ns[0] == ENDN || size == 2 && ns[0] == ENDR { //空行，说明一个数据包结束
		return "", nil, true
	}
	blockSize := ToNum(ns)
	if ns[size-1] == ENDR {
		size = 2
	} else {
		size = 1
	}
	if s.packetBuf.Len() < blockSize+size {
		return "", io.EOF, false
	}

	ns = s.packetBuf.Next(blockSize)
	s.packetBuf.Next(size)
	return string(ns), nil, false
}
