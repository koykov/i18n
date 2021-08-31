package i18n

// Simple entry index map.
type index map[uint64]entry

func (i *index) set(key uint64, lo, hi uint32) {
	var e entry
	e.join(lo, hi)
	(*i)[key] = e
}

func (i index) get(key uint64) entry {
	if e, ok := i[key]; ok {
		return e
	}
	return 0
}
