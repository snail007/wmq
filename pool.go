package main

import (
	"fmt"
	"sync"
)

type poolConfig struct {
	Factory    func() (interface{}, error)
	IsActive   func(interface{}) bool
	Release    func(interface{})
	InitialCap int
	MaxCap     int
}

func newNetPool(poolConfig poolConfig) (pool netPool, err error) {
	pool = netPool{
		config: poolConfig,
		conns:  make(chan interface{}, poolConfig.MaxCap),
		lock:   &sync.Mutex{},
	}
	for i := 0; i < poolConfig.InitialCap; i++ {
		c, err := pool.config.Factory()
		if err != nil {
			err = fmt.Errorf("factory is not able to fill the pool: %s", err)
			pool.ReleaseAll()
			break
		}
		pool.conns <- c
	}
	return
}

type netPool struct {
	conns  chan interface{}
	lock   *sync.Mutex
	config poolConfig
}

func (p *netPool) Get() (conn interface{}, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for {
		select {
		case conn = <-p.conns:
			if p.config.IsActive(conn) {

				return
			}
			p.config.Release(conn)
		default:
			conn, err = p.config.Factory()
			if err != nil {
				return nil, err
			}
			return conn, nil
		}
	}
}

func (p *netPool) Put(conn interface{}) {
	if conn == nil {
		return
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.config.IsActive(conn) {
		p.config.Release(conn)
	}
	select {
	case p.conns <- conn:
	default:
		p.config.Release(conn)
	}
}
func (p *netPool) ReleaseAll() {
	p.lock.Lock()
	defer p.lock.Unlock()
	close(p.conns)
	for c := range p.conns {
		p.config.Release(c)
	}
	p.conns = make(chan interface{}, p.config.InitialCap)

}
func (p *netPool) Len() (length int) {
	return len(p.conns)
}
