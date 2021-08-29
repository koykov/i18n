package i18n

// Simple entry index map.
type index map[uint64]int

func (i *index) set(key uint64, idx int) {
	(*i)[key] = idx
}

func (i index) get(key uint64) int {
	if i, ok := i[key]; ok {
		return i
	}
	return -1
}
