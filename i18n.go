package i18n

import (
	"github.com/koykov/byteptr"
	"github.com/koykov/policy"
)

type DB struct {
	policy.RWLock
	index  index
	locale []*Locale
}

type Locale struct {
	index index
	entry []byteptr.Byteptr
	data  []byte
}

func New() *DB {
	db := &DB{index: make(index)}
	return db
}

func (db *DB) Set(locale, key, translation string) {
	db.Lock()
	defer db.Unlock()
	db.setLF(locale, key, translation)
}

func (db *DB) Get(locale, key string) string {
	var li, ei int
	if li = db.index.get(locale); li == -1 {
		return ""
	}
	loc := db.locale[li]
	if ei = loc.index.get(key); ei == -1 {
		return ""
	}
	return loc.entry[ei].TakeAddr(loc.data).String()
}

func (db *DB) setLF(locale, key, translation string) {
	var li, ei int
	if li = db.index.get(locale); li == -1 {
		loc := &Locale{index: make(index)}
		db.locale = append(db.locale, loc)
		li = len(db.locale) - 1
		db.index[locale] = li
	}
	loc := db.locale[li]
	if ei = loc.index.get(key); ei == -1 {
		offset := len(loc.data)
		loc.data = append(loc.data, translation...)
		ptr := byteptr.Byteptr{}
		ptr.Init(loc.data, offset, len(translation))
		loc.entry = append(loc.entry, ptr)
		loc.index[key] = len(loc.entry) - 1
	} else {
		ptr := &loc.entry[ei]
		if ptr.String() == translation {
			return
		}
		if ptr.Len() >= len(translation) {
			copy(loc.data[ptr.Offset():], translation)
			ptr.SetLen(len(translation))
			return
		}
		offset := len(loc.data)
		loc.data = append(loc.data, translation...)
		ptr.Init(loc.data, offset, len(translation))
	}
}
