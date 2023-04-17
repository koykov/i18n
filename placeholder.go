package i18n

import (
	"github.com/koykov/batch_replace"
	"github.com/koykov/bytealg"
	"github.com/koykov/byteptr"
)

// PlaceholderReplacer is a storage of placeholders replacer.
type PlaceholderReplacer struct {
	kv  []kv
	buf []byte
	br  batch_replace.BatchReplace
}

// Simple key-value pair.
type kv struct {
	k, v byteptr.Byteptr
}

// AddKV stores new placeholder and replace strings as key-value pair.
func (r *PlaceholderReplacer) AddKV(key, value string) *PlaceholderReplacer {
	offsetK := len(r.buf)
	r.buf = append(r.buf, key...)
	bpK := byteptr.Byteptr{}
	bpK.Init(r.buf, offsetK, len(key))

	offsetV := len(r.buf)
	r.buf = append(r.buf, value...)
	bpV := byteptr.Byteptr{}
	bpV.Init(r.buf, offsetV, len(value))

	r.kv = append(r.kv, kv{k: bpK, v: bpV})

	return r
}

// AddSolidKV stores new placeholder and replace string as key-value pair in solid format "<placeholder>:<replace>".
func (r *PlaceholderReplacer) AddSolidKV(kv string) *PlaceholderReplacer {
	offset, i := 0, 0
loop:
	if i = bytealg.IndexAt[string](kv, ":", offset); i == -1 {
		return r
	}
	if i > 0 && kv[i-1] == '\\' {
		goto loop
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
	l := len(r.kv)
	if l == 0 {
		return raw
	}
	r.br.SetSrcStr(raw)
	_ = r.kv[l-1]
	for i := 0; i < l; i++ {
		kv := &r.kv[i]
		r.br.S2S(kv.k.TakeAddr(r.buf).String(), kv.v.TakeAddr(r.buf).String())
	}
	return r.br.CommitStr()
}

// Reset all internal data.
func (r *PlaceholderReplacer) Reset() {
	r.kv = r.kv[:0]
	r.buf = r.buf[:0]
	r.br.Reset()
}
