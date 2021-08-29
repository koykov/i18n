package i18n

import "sync"

// Simple transaction pool.
//
// Have wrappers of Get()/Put() methods.
type txnPool struct {
	sync.Pool
}

var (
	txnP txnPool
)

func (p *txnPool) get() *txn {
	v := p.Pool.Get()
	if v != nil {
		if t, ok := v.(*txn); ok {
			return t
		}
	}
	return &txn{}
}

func (p *txnPool) put(x *txn) {
	x.reset()
	p.Pool.Put(x)
}
