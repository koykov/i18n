package i18n

import (
	"github.com/koykov/batch_replace"
	"github.com/koykov/byteconv"
	"github.com/koykov/byteptr"
	"github.com/koykov/simd/indexbyte"
)

// PlaceholderReplacer is a storage of placeholders replacer.
type PlaceholderReplacer struct {
	kv  []kv
	kvl int
	buf []byte
	br  batch_replace.BatchReplace
}

// Simple key-value pair.
type kv struct {
	k, v byteptr.Byteptr
}

// AddKV stores new placeholder and replace strings as key-value pair.
func (r *PlaceholderReplacer) AddKV(key, value string) *PlaceholderReplacer {
	var bpK, bpV *byteptr.Byteptr
	if r.kvl < len(r.kv) {
		bpK, bpV = &r.kv[r.kvl].k, &r.kv[r.kvl].v
	} else {
		r.kv = append(r.kv, kv{})
		bpK, bpV = &r.kv[len(r.kv)-1].k, &r.kv[len(r.kv)-1].v
	}
	r.kvl++

	offsetK := len(r.buf)
	r.buf = append(r.buf, key...)
	bpK.Init(r.buf, offsetK, len(key))

	offsetV := len(r.buf)
	r.buf = append(r.buf, value...)
	bpV.Init(r.buf, offsetV, len(value))

	return r
}

// AddSolidKV stores new placeholder and replace string as key-value pair in solid format "<placeholder>:<replace>".
func (r *PlaceholderReplacer) AddSolidKV(kv string) *PlaceholderReplacer {
	i := indexbyte.IndexNE(byteconv.S2B(kv), ':')
	if i < 0 {
		return r
	}
	if i+1 >= len(kv) {
		return r
	}
	k, v := kv[:i], kv[i+1:]
	return r.AddKV(k, v)
}

// Size gets count of added replacements.
func (r *PlaceholderReplacer) Size() int {
	return len(r.kv)
}

// Commit performs the replaces.
func (r *PlaceholderReplacer) Commit(raw string) string {
	l := r.kvl
	if l == 0 {
		return raw
	}
	r.br.SetSourceString(raw)
	_ = r.kv[l-1]
	for i := 0; i < l; i++ {
		kv := &r.kv[i]
		r.br.S2S(kv.k.TakeAddress(r.buf).String(), kv.v.TakeAddress(r.buf).String())
	}
	return r.br.CommitString()
}

// Reset all internal data.
func (r *PlaceholderReplacer) Reset() {
	r.kvl = 0
	r.buf = r.buf[:0]
	r.br.Reset()
}
