package ssdbclient

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

// SSDBClient ssdb client
type connection struct {
	//连接写缓冲，默认为8k，单位为kb
	writeBufferSize int
	//连接读缓冲，默认为8k，单位为kb
	readBufferSize int
	//写超时
	writeTimeout int
	//读超时
	readTimeout int
	//创建连接的超时时间，单位为秒。默认值: 5
	connectTimeout int
	//ssdb port
	port int
	//host ssdb host
	host string
	//connection
	sock *net.TCPConn
	//readBuf
	buf []byte
	//write buf
	bufw *bufio.Writer
	//received data
	rsp []byte
	//pos list
	posList []int
	//current pos
	pos int
	//received data size
	rspLen int
	//next end pos
	nextPos int
	//current data size
	dataSize int
	//end char size
	//The input parameter is converted to [] bytes, which by default is converted to json format
	//and can be modified to use a custom serialization
	//将输入参数成[]byte，默认会转换成json格式,可以修改这个参数以便使用自定义的序列化方式
	encodingFunc func(v interface{}) []byte
	//dialer
	dialer *net.Dialer
}

const delim int = 1

// Start start socket
//
//	@return error that may occur on startup. Return nil if successful startup
//
// 启动连接，并设置读写的缓存
func (c *connection) start() error {
	if c.dialer == nil {
		c.dialer = &net.Dialer{Timeout: time.Second * time.Duration(c.connectTimeout)}
	}
	conn, err := c.dialer.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return err
	}
	sock := conn.(*net.TCPConn)
	err = sock.SetReadBuffer(c.readBufferSize * 1024)
	if err != nil {
		return err
	}
	err = sock.SetWriteBuffer(c.writeBufferSize * 1024)
	if err != nil {
		return err
	}
	c.bufw = bufio.NewWriterSize(sock, c.writeBufferSize*1024*2)
	c.buf = make([]byte, c.readBufferSize*1024)
	c.sock = sock
	return nil
}

// Close close SSDBClient
//
//	@return error that may occur on shutdown. Return nil if successful shutdown
func (c *connection) close() error {
	c.buf = nil
	//received data
	c.rsp = nil
	//pos list
	c.posList = nil
	if c.sock == nil {
		return nil
	}
	return c.sock.Close()
}

// write write to buf
func (c *connection) writeBytes(bs []byte) error {
	lbs := strconv.AppendInt(nil, int64(len(bs)), 10)
	if _, err := c.bufw.Write(lbs); err != nil {
		return err
	}
	if err := c.bufw.WriteByte(endN); err != nil {
		return err
	}
	if _, err := c.bufw.Write(bs); err != nil {
		return err
	}
	if err := c.bufw.WriteByte(endN); err != nil {
		return err
	}
	return nil
}

// send cmd to ssdb
func (c *connection) send(args []interface{}) (err error) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			err = c.writeBytes([]byte(arg))
		case []byte:
			err = c.writeBytes(arg)
		case int:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			err = c.writeBytes(bs)
		case int8:
			err = c.writeBytes([]byte{byte(arg)})
		case int16:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			err = c.writeBytes(bs)
		case int32:
			bs := strconv.AppendInt(nil, int64(arg), 10)
			err = c.writeBytes(bs)
		case int64:
			bs := strconv.AppendInt(nil, arg, 10)
			err = c.writeBytes(bs)
		case uint8:
			err = c.writeBytes([]byte{byte(arg)})
		case uint16:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			err = c.writeBytes(bs)
		case uint32:
			bs := strconv.AppendUint(nil, uint64(arg), 10)
			err = c.writeBytes(bs)
		case uint64:
			bs := strconv.AppendUint(nil, arg, 10)
			err = c.writeBytes(bs)
		case float32:
			bs := strconv.AppendFloat(nil, float64(arg), 'g', -1, 32)
			err = c.writeBytes(bs)
		case float64:
			bs := strconv.AppendFloat(nil, arg, 'g', -1, 64)
			err = c.writeBytes(bs)
		case bool:
			if arg {
				err = c.writeBytes([]byte{'1'})
			} else {
				err = c.writeBytes([]byte{'0'})
			}
		case time.Time:
			bs := strconv.AppendInt(nil, arg.Unix(), 10)
			err = c.writeBytes(bs)
		case time.Duration:
			bs := strconv.AppendInt(nil, arg.Nanoseconds(), 10)
			err = c.writeBytes(bs)
		case nil:
			err = c.writeBytes([]byte{})
		default:
			if c.encodingFunc == nil {
				err = errors.New("arguments cannot be serialized, please enable Encoding")
			} else if tbs := c.encodingFunc(arg); tbs != nil {
				err = c.writeBytes(tbs)
			} else {
				err = errors.New("arguments cannot be serialized, please check EncodingFunc")
			}
		}
		if err != nil {
			return err
		}
	}
	if err := c.bufw.WriteByte(endN); err != nil {
		return err
	}
	if err := c.sock.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(c.writeTimeout))); err != nil {
		return err
	}
	if err = c.bufw.Flush(); err != nil {
		return err
	}

	return nil
}

// 第一层，读取数据流
func (c *connection) recv() (resp []string, err error) {
	isEnd := false
	for !isEnd {
		//设置读取数据超时，
		if err = c.sock.SetReadDeadline(time.Now().Add(time.Second * time.Duration(c.readTimeout))); err != nil {
			return nil, err
		}
		n, err := c.sock.Read(c.buf)
		if err != nil {
			return nil, err
		}
		if n < 1 {
			break
		}
		c.rsp = append(c.rsp, c.buf[:n]...)
		c.rspLen += n
		if c.rspLen >= c.nextPos { //保证可以有解析数据
			isEnd, err = c.parseBlock()
			if err != nil {
				return nil, err
			}
			if isEnd {
				max := len(c.posList) - 1
				for i := 0; i < max; i += 2 {
					resp = append(resp, string(c.rsp[c.posList[i]:c.posList[i+1]]))
				}
				break
			}
		}
	}

	c.rsp = nil
	c.rspLen = 0
	c.posList = nil
	c.pos = 0
	c.nextPos = 0
	c.dataSize = 0
	return
}

// 第二层，读取block，直到空行
func (c *connection) parseBlock() (end bool, err error) {
	for c.rspLen >= c.nextPos {
		n := index(c.pos, c.rspLen, '\n', c.rsp)
		if n == -1 {
			break
		}
		end, err = c.parseData(n)
		if err != nil {
			return false, err
		}
		if end { //说明是一个空行
			break
		}
	}
	return
}
func index(pos, size int, c byte, bs []byte) int {
	for i := pos; i < size; i++ {
		if c == bs[i] {
			return i
		}
	}
	return -1
}

// 第三层，读取数据，一个长度一个内容
func (c *connection) parseData(n int) (end bool, err error) {
	if c.dataSize == 0 {
		c.dataSize = toNum(c.rsp[c.pos:n])
		if c.dataSize == 0 {
			return true, nil
		}
		c.nextPos = n + delim + c.dataSize + delim
	}

	if c.nextPos <= c.rspLen {
		c.posList = append(c.posList, n+delim, n+delim+c.dataSize)
		c.pos = n + c.dataSize + delim + 1
		c.dataSize = 0
	}
	return
}
