package gossdb

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/seefan/goerr"
	"net"
	"strconv"
)

type SSDBClient struct {
	isOpen   bool
	Password string
	Host     string
	Port     int
	client   *Client
	sock     *net.TCPConn
	buf      *bufio.Reader
	send_buf bytes.Buffer
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
	sock.SetReadBuffer(1024)
	sock.SetWriteBuffer(1024)

	s.sock = sock

	s.buf = bufio.NewReader(sock)
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
	err := s.send(args)
	if err != nil {
		return nil, err
	}
	resp, err := s.Recv()
	return resp, err
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
		s.sock.Close()
		s.isOpen = false
	}
	return resp, err
}

func (s *SSDBClient) Send(args ...interface{}) error {
	return s.send(args)
}

func (s *SSDBClient) send(args []interface{}) error {
	s.send_buf.Reset()
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.WriteString(arg)
		case []byte:
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case int8:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case uint8:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case uint64:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.send_buf.WriteByte('\n')
			s.send_buf.Write(bs)
		case bool:
			s.send_buf.WriteByte(1)
			s.send_buf.WriteByte('\n')
			if arg {
				s.send_buf.WriteByte(1)
			} else {
				s.send_buf.WriteByte(0)
			}
		case nil:
			s.send_buf.WriteByte(0)
			s.send_buf.WriteByte('\n')
			s.send_buf.WriteString("")
		default:
			if Encoding {
				if bs, err := json.Marshal(arg); err == nil {
					s.send_buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
					s.send_buf.WriteByte('\n')
					s.send_buf.Write(bs)
				} else {
					return fmt.Errorf("bad arguments type,can not json marshal")
				}
			} else {
				return fmt.Errorf("bad arguments type")
			}
		}
		s.send_buf.WriteByte('\n')
	}
	s.send_buf.WriteByte('\n')
	//_, err := s.buf.Write(s.send_buf.Bytes())
	//if err == nil {
	//	err = s.buf.Flush()
	//}
	_, err := s.sock.Write(s.send_buf.Bytes())
	return err
}

func (s *SSDBClient) Recv() ([]string, error) {
	resp := []string{}
	for {
		p, err := s.buf.ReadBytes('\n')
		if err != nil {
			break
		}
		p = p[:len(p)-1]
		if len(p) == 0 || (len(p) == 1 && p[0] == '\r') {
			if len(resp) == 0 {
				continue
			} else {
				return resp, nil
			}
		}
		size := ToNum(p)

		for {
			bs, err := s.buf.Peek(size)
			if err != nil {
				continue
			}
			resp = append(resp, string(bs))
			_, err = s.buf.Discard(size + 1)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return resp, nil
}
