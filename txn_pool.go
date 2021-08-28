package i18n

import "sync"

type txnPool struct {
	sync.Pool
}

var (
	txnP txnPool
)

func (p *txnPool) Get() *TXN {
	v := p.Pool.Get()
	if v != nil {
		if t, ok := v.(*TXN); ok {
			return t
		}
	}
	return &TXN{}
}

func (p *txnPool) Put(x *TXN) {
	x.Reset()
	p.Pool.Put(x)
}
