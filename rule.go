package i18n

import "github.com/koykov/byteptr"

// Rule stores low and high ranges of plural rule and rule's body bytes.
type rule struct {
	lh int64
	bp byteptr.Byteptr
}

// Merge lo/hi ranges and save it.
func (r *rule) encode(lo, hi int32) {
	r.lh = int64(lo)<<32 | int64(hi)
}

// Decode lo/hi ranges.
func (r rule) decode() (lo, hi int32) {
	lo = int32(r.lh >> 32)
	hi = int32(r.lh & 0xffffffff)
	return
}

// Check if count passes the rule's range.
func (r rule) check(count int) bool {
	lo, hi := r.decode()
	return int32(count) >= lo && int32(count) < hi
}
