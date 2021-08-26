package i18n

import (
	"testing"

	"github.com/koykov/policy"
)

func TestDB(t *testing.T) {
	db := New()
	db.SetPolicy(policy.Locked)
	db.Set("en", "messages.welcome", "Hello there!")
	db.SetPolicy(policy.LockFree)
	s := db.Get("en", "messages.welcome")
	if s != "Hello there!" {
		t.Error("translation mismatch")
	}
}
