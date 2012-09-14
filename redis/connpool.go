package redis

import (
	"container/list"
	"fmt"
	"sync"
)

// connPool is a stack-like structure that holds the connections of a Client.
type connPool struct {
	available int
	free      list.List
	lock      sync.Mutex
	emptyCond *sync.Cond
	config    *Config
	closed    bool
}

func newConnPool(config *Config) *connPool {
	cp := &connPool{
		available: config.PoolCapacity,
		config:    config,
		closed:    false,
	}
	fmt.Printf("new pool:%d\n", config.PoolCapacity)

	cp.emptyCond = sync.NewCond(&cp.lock)
	return cp
}

func (cp *connPool) push(c *conn) {

	fmt.Printf("[push]trylock\n")
	cp.lock.Lock()
	fmt.Printf("[push]lock ok a:%d \n", cp.available)

	defer cp.lock.Unlock()
	if cp.closed {
		fmt.Printf("pool is close , so return \n")
		return
	}

	if !c.closed() {
		cp.free.PushFront(c)
	}

	cp.available++

	cp.emptyCond.Signal()

	fmt.Printf("[push]signal ok a:%d \n", cp.available)
}

func (cp *connPool) pull() (c *conn, err *Error) {

	fmt.Printf("[pull]try lock\n")
	cp.lock.Lock()
	fmt.Printf("[pull]lock ok a:%d \n", cp.available)

	defer cp.lock.Unlock()
	if cp.closed {
		return nil, newError("connection pool is closed", ErrorConnection)
	}

	for cp.available == 0 {

		fmt.Printf("[pull]no conn, wait\n")
		cp.emptyCond.Wait()
		fmt.Printf("[pull]wait ok \n")

	}

	if cp.free.Len() > 0 {
		c, _ = cp.free.Remove(cp.free.Back()).(*conn)
	} else {
		// Lazy creation of a connection
		c, err = newConn(cp.config)

		if err != nil {
			fmt.Printf("[pull]conn ok but connct failed a:%d \n", cp.available)
			return nil, err
		}
	}
	cp.available--
	fmt.Printf("[pull]conn ok a:%d \n", cp.available)

	return c, nil
}

func (cp *connPool) close() {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	for e := cp.free.Front(); e != nil; e = e.Next() {
		c, _ := e.Value.(*conn)
		c.close()
	}

	cp.free.Init()
	cp.closed = true
}
