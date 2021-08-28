package i18n

import (
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
	hkey := db.hasher.Sum64(key)

	db.Lock()
	defer db.Unlock()
	db.setLF(hkey, translation)
}

func (db *DB) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	hkey := db.hasher.Sum64(key)

	db.Lock()
	defer db.Unlock()

	var i int
	if i = db.index.get(hkey); i == -1 {
		return ""
	}
	return db.entry[i].TakeAddr(db.data).String()
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
