package i18n

import (
	"github.com/koykov/byteptr"
	"github.com/koykov/policy"
)

type txn struct {
	db   *DB
	buf  []txnRecord
	data []byte
}

type txnRecord struct {
	hkey        uint64
	translation byteptr.Byteptr
}

func (t *txn) set(key, translation string) {
	if t.db == nil {
		return
	}
	hkey := t.db.hasher.Sum64(key)
	if old := t.db.getLF(hkey); old == translation {
		return
	}

	offset := len(t.data)
	t.data = append(t.data, translation...)
	bp := byteptr.Byteptr{}
	bp.Init(t.data, offset, len(translation))
	t.buf = append(t.buf, txnRecord{
		hkey:        hkey,
		translation: bp,
	})
}

func (t *txn) commit() {
	if t.db == nil || len(t.buf) == 0 {
		return
	}
	t.db.SetPolicy(policy.Locked)
	t.db.Lock()

	_ = t.buf[len(t.buf)-1]
	for i := 0; i < len(t.buf); i++ {
		entry := &t.buf[i]
		t.db.setLF(entry.hkey, entry.translation.String())
	}

	t.db.Unlock()
	t.db.SetPolicy(policy.LockFree)

	txnP.Put(t)
}

func (t txn) size() int {
	return len(t.buf)
}

func (t *txn) reset() {
	t.db = nil
	t.buf = t.buf[:0]
	t.data = t.data[:0]
}
