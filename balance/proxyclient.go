package balance

import "github.com/seefan/gossdb/client"

type ProxyClient struct {
	pool *BalancePool
}

func NewProxyClient(pool *BalancePool) *ProxyClient {
	return &ProxyClient{
		pool: pool,
	}
}
func (p *ProxyClient) Do(args ...interface{}) ([]string, error) {
	pc, id, err := p.pool.Get(args)
	if err != nil {
		return nil, err
	}
	defer p.pool.Set(pc, id)
	conn := pc.Client.(client.IClient)
	return conn.Do(args...)
}
