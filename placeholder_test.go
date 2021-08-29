package i18n

import (
	"testing"

	"github.com/koykov/hash/fnv"
	"github.com/koykov/policy"
)

func TestPlaceholderReplacer(t *testing.T) {
	db, _ := New(fnv.Hasher{})
	db.SetPolicy(policy.Locked)
	db.Set("en.user.balance", "Balance of !user: !val !cur")
	db.SetPolicy(policy.LockFree)

	repl := PlaceholderReplacer{}
	repl.AddKV("!user", "John Ruth").
		AddSolidKV("!val:8000").
		AddKV("!cur", "USD")

	s := db.GetWR("en.user.balance", "", &repl)
	if s != "Balance of John Ruth: 8000 USD" {
		t.Errorf("replace mismatch, need '%s', got '%s'", "Balance of John Ruth: 8000 USD", s)
	}
}

func BenchmarkPlaceholderReplacer(b *testing.B) {
	origin, expect := "Balance of !user: !val !cur", "Balance of John Ruth: 8000 USD"

	db, _ := New(fnv.Hasher{})
	db.SetPolicy(policy.Locked)
	db.Set("en.user.balance", origin)
	db.SetPolicy(policy.LockFree)
	repl := PlaceholderReplacer{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		repl.Reset()
		repl.AddKV("!user", "John Ruth").
			AddSolidKV("!val:8000").
			AddKV("!cur", "USD")

		s := db.GetWR("en.user.balance", "", &repl)
		if s != expect {
			b.Errorf("replace mismatch, need '%s', got '%s'", expect, s)
		}
	}
}
