package client

type IClient interface {
	Do(args ...interface{}) ([]string, error)
}
