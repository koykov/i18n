package i18n

import (
	"testing"

	"github.com/koykov/hash/fnv"
	"github.com/koykov/policy"
)

func TestIO(t *testing.T) {
	db, _ := New(fnv.Hasher{})
	db.SetPolicy(policy.Locked)
	db.Set("en.messages.welcome", "Hello there!")
	db.SetPolicy(policy.LockFree)
	s := db.Get("en.messages.welcome")
	if s != "Hello there!" {
		t.Error("translation mismatch")
	}
}
