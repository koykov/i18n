package i18n

import (
	"github.com/koykov/byteptr"
)

// i18n transaction.
type txn struct {
	// Database to apply changes.
	db *DB
	// List of records to apply.
	log []txnRecord
	// Transaction storage.
	buf []byte
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

	offset := len(t.buf)
	t.buf = append(t.buf, translation...)
	bp := byteptr.Byteptr{}
	bp.Init(t.buf, offset, len(translation))
	t.log = append(t.log, txnRecord{
		hkey:        hkey,
		translation: bp,
	})
}

// Apply all transaction changes at once.
//
// Database must be locked.
func (t *txn) commit() {
	if t.db == nil || len(t.log) == 0 {
		return
	}

	_ = t.log[len(t.log)-1]
	for i := 0; i < len(t.log); i++ {
		entry := &t.log[i]
		t.db.setLF(entry.hkey, entry.translation.String())
	}
}

// Get count of collected records.
func (t txn) size() int {
	return len(t.log)
}

// Reset transaction data.
func (t *txn) reset() {
	t.db = nil
	t.log = t.log[:0]
	t.buf = t.buf[:0]
}
