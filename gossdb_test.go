package gossdb

import (
	"testing"
	"time"
	"runtime"
)

func Test_1(t *testing.T) {
	runtime.
	conn, err := NewPool(&Config{
		Host:        "127.0.0.1",
		Port:        6380,
		MinPoolSize: 5,
		MaxPoolSize: 50,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	for i := 0; i < 100; i++ {
		go func() {
			t.Log(conn.ActiveCount, conn.Size, conn.cfg.MaxPoolSize, conn.cfg.AcquireIncrement)
			c, err := conn.NewClient()
			if err != nil {
				t.Error(err)
			}
			time.Sleep(time.Second)
			c.Close()
		}()
	}
	time.Sleep(time.Second * 5)
}
