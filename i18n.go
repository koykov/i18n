package i18n

import (
	"unsafe"

	"github.com/koykov/byteptr"
	"github.com/koykov/hash"
	"github.com/koykov/policy"
)

// i18n database.
type DB struct {
	policy.RWLock
	// Keys hasher.
	hasher hash.Hasher
	// Translations index.
	index index
	// Translations pointer.
	entry []byteptr.Byteptr
	// Translations storage.
	data []byte
	// Transaction pointer.
	txn unsafe.Pointer
}

// Make new DB.
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

// Set translation as key.
//
// If locale needed, the key must contain it as a prefix, eg: "en.messages.accessDenied" or "ru-RU.messages.welcome".
func (db *DB) Set(key, translation string) {
	if len(key) == 0 || len(translation) == 0 {
		return
	}

	db.Lock()
	defer db.Unlock()
	if txn := db.txnIndir(); txn != nil {
		// Save translation to transaction.
		txn.set(key, translation)
	} else {
		// Set transaction immediately.
		hkey := db.hasher.Sum64(key)
		db.setLF(hkey, translation)
	}
}

// Lock-free inner setter.
func (db *DB) setLF(hkey uint64, translation string) {
	var i int
	if i = db.index.get(hkey); i == -1 {
		// Save new translation.
		offset := len(db.data)
		db.data = append(db.data, translation...)
		bp := byteptr.Byteptr{}
		bp.Init(db.data, offset, len(translation))
		db.entry = append(db.entry, bp)
		db.index[hkey] = len(db.entry) - 1
	} else {
		// Update existing translation.
		bp := &db.entry[i]
		if bp.String() == translation {
			// Translation already exists.
			return
		}
		if bp.Len() >= len(translation) {
			// Overwrite translation.
			copy(db.data[bp.Offset():], translation)
			bp.SetLen(len(translation))
			return
		}
		// Write translation at the end of the storage.
		offset := len(db.data)
		db.data = append(db.data, translation...)
		bp.Init(db.data, offset, len(translation))
	}
}

// Get translation of key.
func (db *DB) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	hkey := db.hasher.Sum64(key)

	db.RLock()
	defer db.RUnlock()

	return db.getLF(hkey)
}

// Get translation using plural formula.
func (db *DB) GetPlural(key string, count int) string {
	// todo implement me
	return ""
}

// Lock-free inner getter.
func (db *DB) getLF(hkey uint64) string {
	var i int
	if i = db.index.get(hkey); i == -1 {
		return ""
	}
	return db.entry[i].TakeAddr(db.data).String()
}

// Begin new transaction.
//
// All update calls will collect in the transaction until commit.
func (db *DB) BeginTXN() {
	txn := txnP.get()
	txn.db = db
	db.txn = unsafe.Pointer(txn)
}

// Rollback transaction.
func (db *DB) Rollback() {
	if txn := db.txnIndir(); txn != nil {
		txnP.put(txn)
	}
}

// Commit transaction.
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
