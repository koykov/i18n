package i18n

import (
	"unsafe"

	"github.com/koykov/byteptr"
	"github.com/koykov/hash"
	"github.com/koykov/policy"
)

type DB struct {
	policy.RWLock
	hasher hash.Hasher
	index  index
	entry  []byteptr.Byteptr
	data   []byte
	txn    unsafe.Pointer
}

func New(hasher hash.Hasher) (*DB, error) {
	if hasher == nil {
		return nil, ErrNoHasher
	}
	db := &DB{
		hasher: hasher,
		index:  make(index),
	}
	return db, nil
}

func (db *DB) Set(key, translation string) {
	if len(key) == 0 || len(translation) == 0 {
		return
	}

	db.Lock()
	defer db.Unlock()
	if txn := db.txnIndir(); txn != nil {
		txn.set(key, translation)
	} else {
		hkey := db.hasher.Sum64(key)
		db.setLF(hkey, translation)
	}
}

func (db *DB) setLF(hkey uint64, translation string) {
	var i int
	if i = db.index.get(hkey); i == -1 {
		offset := len(db.data)
		db.data = append(db.data, translation...)
		bp := byteptr.Byteptr{}
		bp.Init(db.data, offset, len(translation))
		db.entry = append(db.entry, bp)
		db.index[hkey] = len(db.entry) - 1
	} else {
		bp := &db.entry[i]
		if bp.String() == translation {
			return
		}
		if bp.Len() >= len(translation) {
			copy(db.data[bp.Offset():], translation)
			bp.SetLen(len(translation))
			return
		}
		offset := len(db.data)
		db.data = append(db.data, translation...)
		bp.Init(db.data, offset, len(translation))
	}
}

func (db *DB) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	hkey := db.hasher.Sum64(key)

	db.RLock()
	defer db.RUnlock()

	return db.getLF(hkey)
}

func (db *DB) GetPlural(key string, count int) string {
	// todo implement me
	return ""
}

func (db *DB) getLF(hkey uint64) string {
	var i int
	if i = db.index.get(hkey); i == -1 {
		return ""
	}
	return db.entry[i].TakeAddr(db.data).String()
}

func (db *DB) BeginTXN() {
	txn := txnP.get()
	txn.db = db
	db.txn = unsafe.Pointer(txn)
}

func (db *DB) Rollback() {
	if txn := db.txnIndir(); txn != nil {
		txnP.put(txn)
	}
}

func (db *DB) Commit() {
	if txn := db.txnIndir(); txn != nil {
		db.SetPolicy(policy.Locked)
		db.Lock()
		txn.commit()
		db.Unlock()
		db.SetPolicy(policy.LockFree)
		txnP.put(txn)
	}
}

func (db *DB) txnIndir() *txn {
	if db.txn == nil {
		return nil
	}
	return (*txn)(db.txn)
}
