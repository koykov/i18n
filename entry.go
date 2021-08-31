package i18n

type entry uint64

func (e *entry) join(lo, hi uint32) {
	*e = entry(lo)<<32 | entry(hi)
}

func (e entry) split() (lo, hi uint32) {
	lo = uint32(e >> 32)
	hi = uint32(e & 0xffffffff)
	return
}
