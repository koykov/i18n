package i18n

import (
	"bytes"
	"math"
	"strconv"
	"unsafe"

	"github.com/koykov/byteptr"
	"github.com/koykov/fastconv"
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
	// Rules storage.
	rules []rule
	// Translations storage.
	buf []byte
	// Transaction pointer.
	txn unsafe.Pointer
}

type rule struct {
	lo, hi int32
	bp     byteptr.Byteptr
}

var (
	inf = []byte("*")
)

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
func (db *DB) setLF(hkey uint64, t9n string) {
	var e entry
	if e = db.index.get(hkey); e == 0 {
		// Save new translation.
		offset := len(db.buf)
		db.buf = append(db.buf, t9n...)
		e = db.makeEntry(offset, len(t9n))
		db.index[hkey] = e
	} else {
		// Update existing translation.
		// todo implement me
	}
}

func (db *DB) makeEntry(off, ln int) entry {
	lo, hi := len(db.rules), len(db.rules)
	s := db.buf[off : off+ln]
	var nextPipe, offPipe, offFormula int
	for i := 0; ; i++ {
		var (
			r      rule
			cb, qb bool
		)
		if nextPipe = db.scanUnescByte(s, '|', offPipe); nextPipe == -1 {
			nextPipe = len(s)
		}
		chunk := s[offPipe:nextPipe]
		if chunk[0] == '{' {
			if lo, offCBE, ok := db.checkCB(chunk, 1); ok {
				offFormula = offCBE
				r.lo = lo
				r.hi = r.lo + 1
				cb = true
			}
		}
		if !cb && chunk[0] == '[' {
			if lo, hi, offFPE, ok := db.checkQB(chunk, 1); ok {
				offFormula = offFPE
				r.lo = lo
				r.hi = hi
				qb = true
			}
		}
		if !cb && !qb {
			if i == 0 {
				r.lo, r.hi = 0, 2
			} else {
				r.lo, r.hi = 2, math.MaxInt32
			}
		}
		r.bp.Init(db.buf, off+offPipe+offFormula, nextPipe-offPipe-offFormula)
		db.rules = append(db.rules, r)
		hi++
		offPipe = nextPipe + 1
		if offPipe >= len(s) {
			break
		}
	}

	var e entry
	e.encode(uint32(lo), uint32(hi))
	return e
}

// Get translation of key.
//
// If translation doesn't exists, def will be used instead.
func (db *DB) Get(key, def string) string {
	return db.GetPluralWR(key, def, 1, nil)
}

// Get translation of key with replacer.
//
// If translation doesn't exists, def will be used instead.
// Replacement rules will apply if repl will pass.
func (db *DB) GetWR(key, def string, repl *PlaceholderReplacer) string {
	return db.GetPluralWR(key, def, 1, repl)
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
	raw := db.getLF(hkey, count)
	db.RUnlock()

	if len(raw) == 0 {
		raw = def
	}
	if len(raw) == 0 {
		return ""
	}

	if repl != nil && repl.Size() > 0 {
		return repl.Commit(raw)
	}

	return raw
}

// Lock-free inner getter.
func (db *DB) getLF(hkey uint64, count int) string {
	var e entry
	if e = db.index.get(hkey); e == 0 {
		return ""
	}
	lo, hi := e.decode()
	if rules := db.rules[lo:hi]; len(rules) > 0 {
		var i int
		_ = rules[len(rules)-1]
	loop:
		rule := rules[i]
		if int32(count) >= rule.lo && int32(count) < rule.hi {
			return rule.bp.TakeAddr(db.buf).String()
		}
		i++
		if i < len(rules) {
			goto loop
		}
	}
	return ""
}

// Get raw translation including all plural formula rules.
func (db *DB) getRawLF(hkey uint64) string {
	var e entry
	if e = db.index.get(hkey); e == 0 {
		return ""
	}
	lo, hi := e.decode()
	if rules := db.rules[lo:hi]; len(rules) > 0 {
		bp := byteptr.Byteptr{}
		bp.TakeAddr(db.buf).SetOffset(rules[0].bp.Offset())
		var i, l int
		_ = rules[len(rules)-1]
	loop:
		rule := rules[i]
		l += rule.bp.Len()
		i++
		if i < len(rules) {
			goto loop
		}
		bp.SetLen(l)
		return bp.TakeAddr(db.buf).String()
	}
	return ""
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

// Indirect transaction from raw pointer.
func (db *DB) txnIndir() *txn {
	if db.txn == nil {
		return nil
	}
	return (*txn)(db.txn)
}

// Get next position of unescaped b.
func (db *DB) scanUnescByte(s []byte, b byte, offset int) int {
	for si := bytes.IndexByte(s[offset:], b); si != -1; {
		if si > 0 && s[si-1] == '\\' {
			offset = si
			continue
		}
		return offset + si
	}
	return -1
}

// Check value in curly brackets.
//
// Returns the exact value, offset of rule payload and success flag.
func (db *DB) checkCB(p []byte, off int) (lo int32, offCBE int, ok bool) {
	if offCBE = db.scanUnescByte(p, '}', off); offCBE != -1 {
		if raw := p[off:offCBE]; len(raw) > 0 {
			if lo64, err := strconv.ParseInt(fastconv.B2S(raw), 10, 32); err == nil {
				if p[offCBE+1] == ' ' {
					offCBE += 2
				} else {
					offCBE++
				}
				lo = int32(lo64)
				ok = true
			}
		}
	}
	return
}

// Check values in square brackets.
//
// Returns the low/high values of range, offset of rule payload and success flag.
func (db *DB) checkQB(p []byte, off int) (lo int32, hi int32, offQBE int, ok bool) {
	if offQBE = db.scanUnescByte(p, ']', off); offQBE != -1 {
		raw := p[off:offQBE]
		if p[offQBE+1] == ' ' {
			offQBE += 2
		} else {
			offQBE++
		}
		if offComma := bytes.IndexByte(raw, ','); offComma != -1 {
			rawLo, rawHi := raw[:offComma], raw[offComma+1:]
			ok = true
			if bytes.Equal(rawLo, inf) {
				lo = math.MinInt32
			} else if lo64, err := strconv.ParseInt(fastconv.B2S(rawLo), 10, 32); err == nil {
				lo = int32(lo64)
			} else {
				ok = false
			}
			if bytes.Equal(rawHi, inf) {
				hi = math.MaxInt32
			} else if hi64, err := strconv.ParseInt(fastconv.B2S(rawHi), 10, 32); err == nil {
				hi = int32(hi64)
			} else {
				ok = false
			}
		}
	}
	return
}
