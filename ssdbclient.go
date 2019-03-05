package gossdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/seefan/goerr"
)

const (
	ENDN = '\n'
	ENDR = '\r'
)

type SSDBClient struct {
	isOpen   bool
	Password string
	Host     string
	Port     int
	client   *Client
	sock     *net.TCPConn
	readBuf  []byte
	//packetBuf bytes.Buffer
	//连接写缓冲，默认为8k，单位为kb
	WriteBufferSize int
	//连接读缓冲，默认为8k，单位为kb
	ReadBufferSize int
	//是否重试
	RetryEnabled bool
	//读写超时
	ReadWriteTimeout int
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
	return nil
}

//关闭连接
func (s *SSDBClient) Close() error {
	s.isOpen = false
	s.readBuf = nil
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
		return nil, goerr.New("gossdb client is closed.")
	}
	err := s.send(args)
	if err != nil {
		return nil, goerr.NewError(err, "client send error")
	}
	if resp, err := s.Recv(); err != nil {
		return nil, goerr.NewError(err, "client recv error")
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
				err = goerr.NewError(e, "client close failed")
			}
			return nil, goerr.NewError(err, "authentication failed")
		}
		if len(resp) > 0 && resp[0] == OK {
			//验证成功
			s.Password = ""
		} else {
			return nil, goerr.New("Authentication failed,password is wrong")
		}
	}
	resp, err := s.do(args...)
	if err != nil {
		if e := s.Close(); e != nil {
			err = goerr.NewError(e, "client close failed")
		}
		if s.RetryEnabled { //如果允许重试，就重新打开一次连接
			if err = s.Start(); err == nil {
				resp, err = s.do(args...)
				if err != nil {
					if e := s.Close(); e != nil {
						err = goerr.NewError(e, "client close failed")
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
			err = goerr.NewError(e, "client close failed")
		}
		return err
	}
	return nil
}

//发送数据
func (s *SSDBClient) send(args []interface{}) error {
	var packetBuf bytes.Buffer
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			packetBuf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.WriteString(arg)
		case []string:
			for _, a := range arg {
				packetBuf.Write(strconv.AppendInt(nil, int64(len(a)), 10))
				packetBuf.WriteByte(ENDN)
				packetBuf.WriteString(a)
				packetBuf.WriteByte(ENDN)
			}
			continue
		case []byte:
			packetBuf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case int8:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case uint8:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case uint64:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case bool:
			packetBuf.WriteByte(1)
			packetBuf.WriteByte(ENDN)
			if arg {
				packetBuf.WriteByte(1)
			} else {
				packetBuf.WriteByte(0)
			}
		case time.Time:
			bs := strconv.AppendInt(nil, arg.Unix(), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case time.Duration:
			bs := strconv.AppendInt(nil, arg.Nanoseconds(), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte(ENDN)
			packetBuf.Write(bs)
		case nil:
			packetBuf.WriteByte(0)
			packetBuf.WriteByte(ENDN)
			packetBuf.WriteString("")
		default:
			if Encoding {
				if bs, err := json.Marshal(arg); err == nil {
					packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
					packetBuf.WriteByte(ENDN)
					packetBuf.Write(bs)
				} else {
					return fmt.Errorf("bad arguments type,can not json marshal")
				}
			} else {
				return fmt.Errorf("bad arguments type")
			}
		}
		packetBuf.WriteByte(ENDN)
	}
	packetBuf.WriteByte(ENDN)
	if err := s.sock.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(s.ReadWriteTimeout))); err != nil {
		return err
	}
	for _, err := packetBuf.WriteTo(s.sock); packetBuf.Len() > 0; {
		if err != nil {
			packetBuf.Reset()
			return goerr.NewError(err, "client socket write error")
		}
	}
	//设置不超时
	if err := s.sock.SetWriteDeadline(s.timeZero); err != nil {
		return err
	}
	packetBuf.Reset()
	return nil
}

//接收数据
func (s *SSDBClient) Recv() (resp []string, err error) {
	bufSize := 0
	packetBuf := []byte{}
	//设置读取数据超时，
	if err = s.sock.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ReadWriteTimeout))); err != nil {
		return nil, err
	}
	//数据包分解，发现长度，找到结尾，循环发现，发现空行，结束
	for {
		bufSize, err = s.sock.Read(s.readBuf)
		if err != nil {
			return nil, goerr.NewError(err, "client socket read error")
		}
		if bufSize < 1 {
			continue
		}
		packetBuf = append(packetBuf, s.readBuf[:bufSize]...)

		for {
			rsp, n := s.parse(packetBuf)
			if n == -1 {
				break
			} else if n == -2 {
				return
			} else {
				resp = append(resp, rsp)
				packetBuf = packetBuf[n+1:]
			}
		}
	}
	packetBuf = nil
	//设置不超时
	if err = s.sock.SetReadDeadline(s.timeZero); err != nil {
		return nil, err
	}
	return resp, nil
}

//解析数据为string的slice
func (s *SSDBClient) parse(buf []byte) (resp string, size int) {
	n := bytes.IndexByte(buf, ENDN)
	blockSize := -1
	size = -1
	if n != -1 {
		if n == 0 || n == 1 && buf[0] == ENDR { //空行，说明一个数据包结束
			size = -2
			return
		}
		//数据包开始，包长度解析
		blockSize = ToNum(buf[:n])
		bufSize := len(buf)

		if n+blockSize < bufSize {
			resp = string(buf[n+1 : blockSize+n+1])
			for i := blockSize + n + 1; i < bufSize; i++ {
				if buf[i] == ENDN {
					size = i
					return
				}
			}
		}
	}

	return
}
