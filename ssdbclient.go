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
	isOpen    bool
	Password  string
	Host      string
	Port      int
	client    *Client
	sock      *net.TCPConn
	buf       *bufio.ReadWriter
	packetBuf bytes.Buffer
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
	sock.SetReadBuffer(8192)
	//sock.SetWriteBuffer(1024)
	s.buf = bufio.NewReadWriter(bufio.NewReader(sock), bufio.NewWriter(sock))
	s.sock = sock

	s.isOpen = true
	return nil
}
func (s *SSDBClient) Close() error {
	s.isOpen = false
	s.buf = nil
	s.packetBuf.Reset()
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

	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			s.buf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			s.buf.WriteByte('\n')
			s.buf.WriteString(arg)
		case []byte:
			s.buf.Write(strconv.AppendInt(nil, int64(len(arg)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case int8:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case uint8:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case uint64:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
			s.buf.WriteByte('\n')
			s.buf.Write(bs)
		case bool:
			s.buf.WriteByte(1)
			s.buf.WriteByte('\n')
			if arg {
				s.buf.WriteByte(1)
			} else {
				s.buf.WriteByte(0)
			}
		case nil:
			s.buf.WriteByte(0)
			s.buf.WriteByte('\n')
			s.buf.WriteString("")
		default:
			if Encoding {
				if bs, err := json.Marshal(arg); err == nil {
					s.buf.Write(strconv.AppendInt(nil, int64(len(bs)), 10))
					s.buf.WriteByte('\n')
					s.buf.Write(bs)
				} else {
					return fmt.Errorf("bad arguments type,can not json marshal")
				}
			} else {
				return fmt.Errorf("bad arguments type")
			}
		}
		s.buf.WriteByte('\n')
	}
	s.buf.WriteByte('\n')
	return s.buf.Flush()
}

func (s *SSDBClient) Recv() (resp []string, err error) {
	packetSize := -1
	drop := 1
	bufSize := 0
	s.packetBuf.Reset()
	for {
		buf, err := s.buf.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		bufSize = len(buf)
		if packetSize == -1 && (bufSize == 1 || bufSize == 2 && buf[0] == '\r') { //空行，说明一个数据包结束
			return resp, nil
		}
		if bufSize > 2 && buf[bufSize-2] == '\r' { // drop end
			drop = 2
		} else {
			drop = 1
		}
		if packetSize == -1 {
			packetSize = ToNum(buf[:(bufSize - drop)])
		} else {
			if s.packetBuf.Len()+bufSize == packetSize+drop {
				s.packetBuf.Write(buf[:bufSize-drop])
				resp = append(resp, s.packetBuf.String())
				s.packetBuf.Reset()
				packetSize = -1
			} else {
				s.packetBuf.Write(buf)
			}
		}
	}
	return resp, nil
}
