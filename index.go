package i18n

import "github.com/koykov/entry"

// Index stores hashed key-entry pairs.
//
// Hashed key uses to reduce pointers in the package to follow pointers policy.
// Entry as uint64 value uses due to impossibility to take a pointer of map value.
type index map[uint64]entry.Entry64

// Save new entry.
func (i *index) set(key uint64, lo, hi uint32) {
	var e entry.Entry64
	e.Encode(lo, hi)
	(*i)[key] = e
}

// Get entry by given key.
func (i index) get(key uint64) entry.Entry64 {
	if e, ok := i[key]; ok {
		return e
	}
	return 0
}

// Remove all keys from index.
func (i *index) reset() {
	for h := range *i {
		delete(*i, h)
	}
}
