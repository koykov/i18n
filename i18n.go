package i18n

import (
	"strings"
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
	buf []byte
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
		offset := len(db.buf)
		db.buf = append(db.buf, translation...)
		bp := byteptr.Byteptr{}
		bp.Init(db.buf, offset, len(translation))
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
			copy(db.buf[bp.Offset():], translation)
			bp.SetLen(len(translation))
			return
		}
		// Write translation at the end of the storage.
		offset := len(db.buf)
		db.buf = append(db.buf, translation...)
		bp.Init(db.buf, offset, len(translation))
	}
}

// Get translation of key.
//
// If translation doesn't exists, def will be used instead.
func (db *DB) Get(key, def string) string {
	return db.GetWR(key, def, nil)
}

// Get translation of key with replacer.
//
// If translation doesn't exists, def will be used instead.
// Replacement rules will apply if repl will pass.
func (db *DB) GetWR(key, def string, repl *PlaceholderReplacer) string {
	if len(key) == 0 {
		return ""
	}
	hkey := db.hasher.Sum64(key)

	db.RLock()
	raw := db.getLF(hkey)
	db.RUnlock()

	if len(raw) == 0 {
		raw = def
	}
	if len(raw) != 0 && repl != nil {
		return repl.Commit(raw)
	}
	return raw
}

// Get translation using plural formula.
func (db *DB) GetPlural(key, def string, count int) string {
	return db.GetPluralWR(key, def, count, nil)
}

// Get translation using plural formula with replacer.
//
// See GetWR().
func (db *DB) GetPluralWR(key, def string, count int, repl *PlaceholderReplacer) string {
	if len(key) == 0 {
		return ""
	}
	hkey := db.hasher.Sum64(key)

	db.RLock()
	raw := db.getLF(hkey)
	db.RUnlock()

	if len(raw) == 0 {
		raw = def
	}
	if len(raw) == 0 {
		return ""
	}

	// Handle separators.
	var r string
	prim, offset, poffset := false, 0, 0
	for pos := db.scanUnescByte(raw, '|', offset); pos != -1; {
		poffset = offset
		offset = pos
		if pos+1 < len(raw) {
			switch raw[pos+1] {
			case '{':
				pos1 := db.scanUnescByte(raw, '}', pos+1)
				_ = pos1
				// todo parse exact plural rule
			case '[':
				pos1 := db.scanUnescByte(raw, ']', pos+1)
				_ = pos1
				// todo parse range plural rule
			default:
				prim = true
				break
			}
		}
	}
	if prim {
		switch count {
		case 1:
			r = raw[:offset]
		default:
			r = raw[offset+1:]
		}
	} else {
		_ = poffset
		// todo handle exact/range plural rule
	}

	if len(r) > 0 && repl != nil {
		return repl.Commit(r)
	}

	return r
}

// Lock-free inner getter.
func (db *DB) getLF(hkey uint64) string {
	var i int
	if i = db.index.get(hkey); i == -1 {
		return ""
	}
	return db.entry[i].TakeAddr(db.buf).String()
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

func (db *DB) scanUnescByte(s string, b byte, offset int) int {
	for si := strings.IndexByte(s[offset:], b); si != -1; {
		offset = si
		if si > 0 && s[si-1] == '\\' {
			continue
		}
		return si
	}
	return -1
}
