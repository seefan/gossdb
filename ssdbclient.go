package gossdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/seefan/goerr"
)

type SSDBClient struct {
	isOpen   bool
	Password string
	Host     string
	Port     int
	client   *Client
	sock     *net.TCPConn

	//packetBuf bytes.Buffer
	//连接写缓冲，默认为8k，单位为kb
	WriteBufferSize int
	//连接读缓冲，默认为8k，单位为kb
	ReadBufferSize int
	//是否重试
	RetryEnabled bool
}

//打开连接
func (s *SSDBClient) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return err
	}
	sock, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	sock.SetReadBuffer(s.ReadBufferSize * 1024)
	sock.SetWriteBuffer(s.WriteBufferSize * 1024)
	s.sock = sock
	s.isOpen = true
	return nil
}
func (s *SSDBClient) Close() error {
	s.isOpen = false
	return s.sock.Close()
}
func (s *SSDBClient) IsOpen() bool {

	return s.isOpen
}
func (s *SSDBClient) Ping() bool {
	_, err := s.Do("info")
	return err == nil
}
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
			s.sock.Close()
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
	resp, err := s.do(args...)
	if err != nil {
		s.Close()
		if s.RetryEnabled { //如果允许重试，就重新打开一次连接
			if err = s.Start(); err == nil {
				resp, err = s.do(args...)
				if err != nil {
					s.Close()
				}
			}
		}
	}
	return resp, err
}

func (s *SSDBClient) Send(args ...interface{}) error {
	return s.send(args)
}

func (s *SSDBClient) send(args []interface{}) error {
	var packetBuf bytes.Buffer
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			packetBuf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.WriteString(arg)
		case []string:
			for _, a := range arg {
				packetBuf.Write(strconv.AppendInt(nil, int64(len(a)), 10))
				packetBuf.WriteByte('\n')
				packetBuf.WriteString(a)
				packetBuf.WriteByte('\n')
			}
			continue
		case []byte:
			packetBuf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case int8:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case uint8:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case uint64:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			packetBuf.WriteByte('\n')
			packetBuf.Write(bs)
		case bool:
			packetBuf.WriteByte(1)
			packetBuf.WriteByte('\n')
			if arg {
				packetBuf.WriteByte(1)
			} else {
				packetBuf.WriteByte(0)
			}
		case nil:
			packetBuf.WriteByte(0)
			packetBuf.WriteByte('\n')
			packetBuf.WriteString("")
		default:
			if Encoding {
				if bs, err := json.Marshal(arg); err == nil {
					packetBuf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
					packetBuf.WriteByte('\n')
					packetBuf.Write(bs)
				} else {
					return fmt.Errorf("bad arguments type,can not json marshal")
				}
			} else {
				return fmt.Errorf("bad arguments type")
			}
		}
		packetBuf.WriteByte('\n')
	}
	packetBuf.WriteByte('\n')

	for _, err := packetBuf.WriteTo(s.sock); packetBuf.Len() > 0; {
		if err != nil {
			return goerr.NewError(err, "client socket write error")
		}
	}
	return nil
}

func (s *SSDBClient) Recv() (resp []string, err error) {
	bufSize := 0
	packetBuf := []byte{}
	buf := make([]byte, 1024)
	//数据包分解，发现长度，找到结尾，循环发现，发现空行，结束
	for {
		bufSize, err = s.sock.Read(buf)
		if err != nil {
			return nil, goerr.NewError(err, "client socket read error")
		}
		if bufSize < 1 {
			continue
		}
		packetBuf = append(packetBuf, buf[:bufSize]...)

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
	return resp, nil
}
func (s *SSDBClient) parse(buf []byte) (resp string, size int) {
	n := bytes.IndexByte(buf, '\n')
	blockSize := -1
	size = -1
	if n != -1 {
		if n == 0 || n == 1 && buf[0] == '\r' { //空行，说明一个数据包结束
			size = -2
			return
		}
		//数据包开始，包长度解析
		blockSize = ToNum(buf[:n])
		bufSize := len(buf)

		if n+blockSize < bufSize {
			resp = string(buf[n+1 : blockSize+n+1])
			for i := blockSize + n + 1; i < bufSize; i++ {
				if buf[i] == '\n' {
					size = i
					return
				}
			}
		}
	}

	return
}
