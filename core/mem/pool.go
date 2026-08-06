package mem

import (
	"sync"
)

type Pool struct {
	c    chan any
	pool sync.Pool
}

func NewPool(c chan any, new func() any) *Pool {
	return &Pool{
		c:    c,
		pool: sync.Pool{New: new},
	}
}

func (m *Pool) Put(data any) {
	select {
	case m.c <- data:
	default:
		m.pool.Put(data)
	}
}

func (m *Pool) Get() any {
	select {
	case data := <-m.c:
		return data
	default:
		return m.pool.Get()
	}
}
