package symc

import (
	"container/list"
	"errors"
	"sync"
)

type Pool struct {
	vbucketConf string
	vbucketName string
	pool        *list.List
	lock        sync.Mutex
}

func NewSymcPool(vbucketConf, vbucketName string) (*Pool, error) {
	s, err := NewSymc(vbucketConf, vbucketName)
	if err != nil {
		return nil, err
	}
	p := &Pool{
		vbucketConf: vbucketConf,
		vbucketName: vbucketName,
	}
	p.pool = list.New()
	p.lock.Lock()
	defer p.lock.Unlock()
	p.pool.PushBack(s)
	s.p = p
	return p, nil
}

func (p *Pool) GetSymc() (*Symc, error) {
	if p.pool == nil {
		return nil, errors.New("pool is nil")
	}
	if p.pool.Len() == 0 {
		s, err := NewSymc(p.vbucketConf, p.vbucketName)
		if err != nil {
			return nil, err
		}
		s.p = p
		return s, nil
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	sE := p.pool.Front()
	p.pool.Remove(sE)
	return sE.Value.(*Symc), nil
}

func (p *Pool) returnSymc(s *Symc) {
	if s == nil {
		return
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	p.pool.PushBack(s)
}

func (p *Pool) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for e := p.pool.Front(); e != nil; e = e.Next() {
		s := e.Value.(*Symc)
		s.destory()
	}

	p.pool = nil
}
