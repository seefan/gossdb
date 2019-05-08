package ssdb

import (
	"testing"

	"github.com/seefan/gossdb"
	"github.com/seefan/gossdb/conf"
)

func TestStart(t *testing.T) {
	type args struct {
		cfgList []*conf.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{cfgList: []*conf.Config{{Host: "127.0.0.1", Port: 8888}}}, false},
		{"2", args{cfgList: []*conf.Config{{Host: "127.0.0.1", Port: 8889}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Start(tt.args.cfgList...); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClose(t *testing.T) {
	err := Start(&conf.Config{
		MaxPoolSize: 10,
		MinPoolSize: 10,
		PoolSize:    10,
		Host:        "127.0.0.1",
		Port:        8888,
	})
	if err != nil {
		t.Error(err)
	}
	defer Close()
	for i := 0; i < 100; i++ {
		if v, err := ClientAutoClose().Get("a"); err != nil {
			t.Error(err)
		} else {
			t.Log(v)
		}
	}
}

func TestSimple(t *testing.T) {
	err := Start(&conf.Config{
		AutoClose:   true,
		MaxPoolSize: 10,
		MinPoolSize: 10,
		PoolSize:    10,
		Host:        "127.0.0.1",
		Port:        8888,
	})
	if err != nil {
		t.Error(err)
	}
	defer Close()
	err = Simple(func(c *gossdb.Client) error {
		if v, err := c.Get("a"); err != nil {
			return err
		} else {
			t.Log(v)
		}
		return nil

	})
	if err != nil {
		t.Error()
	}
}
