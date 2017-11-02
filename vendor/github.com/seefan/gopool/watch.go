package gopool

var (
	//检查的位置
	checkIndex = -1
	//当前时间
	now int64
)
//watch spare element
func (p *Pool) watch() {
	timeOut := int64(p.IdleTime)
	for t := range p.watcher.C {
		now = t.Unix()
		if p.waitCount == 0 && p.Status == PoolStart {
			if p.pooled.length <= p.MinPoolSize {
				p.pooled.checkMinPoolClient(timeOut)
				if p.pooled.length == 0 {
					p.Status = PoolReStart
					if err := p.Start(); err == nil {
						p.Status = PoolStart
					}
				}
			} else {
				p.pooled.checkPool(timeOut)
			}
		}
	}
}

//检查最小连接池以外的连接，current以外的连接如果不用就回收
//
// hs，int64，超时时间
func (s *Slice) checkPool(hs int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	pos := s.length - 1
	if s.length > s.minPoolSize && pos > s.current && s.pooled[pos] != nil && !s.pooled[pos].isUsed && s.pooled[pos].lastTime+hs < now {
		s.pooled[pos].Client.Close()
		s.length -= 1
	}
}

//检查最小连接池以内的连接，如果不用就ping下，以保持连接一直有数据，如果ping不能，就重启下。重启不成功就关掉。
//
// hs，int64，超时时间
func (s *Slice) checkMinPoolClient(hs int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if checkIndex < 0 || checkIndex < s.current {
		checkIndex = s.length - 1
	}
	//同一个连接检查要间隔HealthSecond秒
	if s.pooled[checkIndex] != nil && !s.pooled[checkIndex].isUsed && s.pooled[checkIndex].lastTime+hs < now {
		s.pooled[checkIndex].lastTime = now
		if !s.pooled[checkIndex].Client.Ping() {
			s.pooled[checkIndex].Client.Close()
			if err := s.pooled[checkIndex].Client.Start(); err != nil {
				s.length -= 1
			}
		}
	}
	checkIndex -= 1
}
