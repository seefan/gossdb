package ssdbclient

import (
	"net"
	"testing"
	"time"
)

func TestSSDBClient_Start(t *testing.T) {
	type fields struct {
		isOpen           bool
		Password         string
		Host             string
		Port             int
		sock             *net.TCPConn
		readBuf          []byte
		WriteBufferSize  int
		ReadBufferSize   int
		RetryEnabled     bool
		ReadWriteTimeout int
		timeZero         time.Time
		ConnectTimeout   int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"start", fields{Host: "127.0.0.1", Port: 8888, ReadBufferSize: 1024, WriteBufferSize: 1024}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSDBClient{
				isOpen:           tt.fields.isOpen,
				Password:         tt.fields.Password,
				host:             tt.fields.Host,
				Port:             tt.fields.Port,
				sock:             tt.fields.sock,
				readBuf:          tt.fields.readBuf,
				WriteBufferSize:  tt.fields.WriteBufferSize,
				ReadBufferSize:   tt.fields.ReadBufferSize,
				RetryEnabled:     tt.fields.RetryEnabled,
				ReadWriteTimeout: tt.fields.ReadWriteTimeout,
				timeZero:         tt.fields.timeZero,
				ConnectTimeout:   tt.fields.ConnectTimeout,
			}
			if err := s.Start(); (err != nil) != tt.wantErr {
				t.Errorf("SSDBClient.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("SSDBClient.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSSDBClient_Ping(t *testing.T) {
	type fields struct {
		isOpen           bool
		Password         string
		Host             string
		Port             int
		sock             *net.TCPConn
		readBuf          []byte
		WriteBufferSize  int
		ReadBufferSize   int
		RetryEnabled     bool
		ReadWriteTimeout int
		timeZero         time.Time
		ConnectTimeout   int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"start", fields{Host: "127.0.0.1", Port: 8888, ReadBufferSize: 1024, WriteBufferSize: 1024, ReadWriteTimeout: 100}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSDBClient{
				isOpen:           tt.fields.isOpen,
				Password:         tt.fields.Password,
				host:             tt.fields.Host,
				Port:             tt.fields.Port,
				sock:             tt.fields.sock,
				readBuf:          tt.fields.readBuf,
				WriteBufferSize:  tt.fields.WriteBufferSize,
				ReadBufferSize:   tt.fields.ReadBufferSize,
				RetryEnabled:     tt.fields.RetryEnabled,
				ReadWriteTimeout: tt.fields.ReadWriteTimeout,
				timeZero:         tt.fields.timeZero,
				ConnectTimeout:   tt.fields.ConnectTimeout,
			}
			err := s.Start()
			if err != nil {
				t.Fatalf("SSDBClient.Start() error = %v", err)
				return
			}
			if got, err := s.Do("get", "a"); err != nil {
				t.Errorf("SSDBClient.Ping() = %v, want %v", got, tt.want)
			}
			for i := 0; i < 100; i++ {
				if got, err := s.Do("get", "a"); err != nil {
					t.Errorf("SSDBClient.get() = %v, want %v", got, tt.want)
				} else if got[0] != "ok" || got[1] != "1" {
					t.Errorf("SSDBClient.get() = %v, want %v", got, tt.want)
				}
				if got, err := s.Do("set", "a", "1"); err != nil {
					t.Errorf("SSDBClient.set() = %v, want %v", got, tt.want)
				}
			}
			if err := s.Close(); err != nil {
				t.Fatalf("SSDBClient.Close() error = %v", err)
			}

		})
	}
}
