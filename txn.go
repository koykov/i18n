package i18n

import (
	"github.com/koykov/byteptr"
)

// i18n transaction.
type txn struct {
	// Database to apply changes.
	db *DB
	// List of records to apply.
	buf []txnRecord
	// Transaction storage.
	data []byte
}

// Key-translation pair of transaction.
type txnRecord struct {
	hkey        uint64
	translation byteptr.Byteptr
}

// Collect new translation.
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

// Apply all transaction changes at once.
//
// Database must be locked.
func (t *txn) commit() {
	if t.db == nil || len(t.buf) == 0 {
		return
	}

	_ = t.buf[len(t.buf)-1]
	for i := 0; i < len(t.buf); i++ {
		entry := &t.buf[i]
		t.db.setLF(entry.hkey, entry.translation.String())
	}

	txnP.Put(t)
}

// Get count of collected records.
func (t txn) size() int {
	return len(t.buf)
}

// Reset transaction data.
func (t *txn) reset() {
	t.db = nil
	t.buf = t.buf[:0]
	t.data = t.data[:0]
}
