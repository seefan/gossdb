package balance

type ProxyClient struct {
	pool *Pool
}

func (p *ProxyClient) Close() error {
	return nil
}
func (p *ProxyClient) Start() error {
	return nil
}
func (p *ProxyClient) IsOpen() bool {
	return true
}
func (p *ProxyClient) Do(args ...interface{}) ([]string, error) {
	return nil, nil
}
