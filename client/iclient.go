package client

import "github.com/seefan/gopool"

type IClient interface {
	gopool.IClient
	Do(args ...interface{}) ([]string, error)
}
