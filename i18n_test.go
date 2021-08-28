package i18n

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/koykov/fastconv"
	"github.com/koykov/hash/fnv"
	"github.com/koykov/policy"
)

func TestIO(t *testing.T) {
	testIO := func(t *testing.T, entries int64) {
		buf := []byte("en.")
		db, _ := New(fnv.Hasher{})
		db.SetPolicy(policy.Locked)
		for i := int64(0); i < entries; i++ {
			buf = strconv.AppendInt(buf[:3], i, 10)
			db.Set(fastconv.B2S(buf), "Hello there!")
		}
		db.SetPolicy(policy.LockFree)

		i := rand.Int63n(entries)
		buf = strconv.AppendInt(buf[:3], i, 10)
		s := db.Get(fastconv.B2S(buf))
		if s != "Hello there!" {
			t.Error("translation mismatch")
		}
	}

	t.Run("1", func(t *testing.T) { testIO(t, 1) })
	t.Run("10", func(t *testing.T) { testIO(t, 10) })
	t.Run("100", func(t *testing.T) { testIO(t, 100) })
	t.Run("1K", func(t *testing.T) { testIO(t, 1000) })
	t.Run("10K", func(t *testing.T) { testIO(t, 10000) })
	t.Run("100K", func(t *testing.T) { testIO(t, 100000) })
}
