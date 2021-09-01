package i18n

// Entry stores low and high indices in rules array.
type entry uint64

// Merge lo/hi indexes and save it.
func (e *entry) encode(lo, hi uint32) {
	*e = entry(lo)<<32 | entry(hi)
}

// Decode lo/hi indexes from the entry.
func (e entry) decode() (lo, hi uint32) {
	lo = uint32(e >> 32)
	hi = uint32(e & 0xffffffff)
	return
}
