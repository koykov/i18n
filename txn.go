package i18n

import (
	"github.com/koykov/byteptr"
	"github.com/koykov/policy"
)

type TXN struct {
	db   *DB
	buf  []txnRecord
	data []byte
}

type txnRecord struct {
	key, translation byteptr.Byteptr
}

func (t *TXN) Set(key, translation string) {
	if t.db == nil {
		return
	}
	if old := t.db.Get(key); old == translation {
		return
	}

	offsetK := len(t.data)
	t.data = append(t.data, key...)
	bpK := byteptr.Byteptr{}
	bpK.Init(t.data, offsetK, len(key))

	offsetT := len(t.data)
	t.data = append(t.data, translation...)
	bpT := byteptr.Byteptr{}
	bpT.Init(t.data, offsetT, len(translation))
	t.buf = append(t.buf, txnRecord{
		key:         bpK,
		translation: bpT,
	})
}

func (t *TXN) Commit() {
	if t.db == nil || len(t.buf) == 0 {
		return
	}
	t.db.SetPolicy(policy.Locked)
	t.db.Lock()

	_ = t.buf[len(t.buf)-1]
	for i := 0; i < len(t.buf); i++ {
		entry := &t.buf[i]
		hkey := t.db.hasher.Sum64(entry.key.String())
		t.db.setLF(hkey, entry.translation.String())
	}

	t.db.Unlock()
	t.db.SetPolicy(policy.LockFree)

	txnP.Put(t)
}

func (t TXN) Size() int {
	return len(t.buf)
}

func (t *TXN) Reset() {
	t.db = nil
	t.buf = t.buf[:0]
	t.data = t.data[:0]
}
