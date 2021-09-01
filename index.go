package i18n

// Index stores hashed key-entry pairs.
//
// Hashed key uses to reduce pointers in the package to follow pointers policy.
// Entry as uint64 value uses due to impossibility tot ake an pointer of map value.
type index map[uint64]entry

// Save new entry.
func (i *index) set(key uint64, lo, hi uint32) {
	var e entry
	e.encode(lo, hi)
	(*i)[key] = e
}

// Get entry by given key.
func (i index) get(key uint64) entry {
	if e, ok := i[key]; ok {
		return e
	}
	return 0
}
