package i18n

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/koykov/fastconv"
	"github.com/koykov/hash/fnv"
	"github.com/koykov/policy"
)

func assertLH(t *testing.T, lo, hi, loE, hiE uint32) {
	if lo != loE || hi != hiE {
		t.Errorf("rules range mismatch, need[%d,%d], got [%d,%d]", loE, hiE, lo, hi)
	}
}
func assertT9n(t *testing.T, db *DB, key, expect string) {
	t9n := db.Get(key, "")
	if t9n != expect {
		t.Errorf("translation mismatch, need %s, got %s", expect, t9n)
	}
}

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
		s := db.Get(fastconv.B2S(buf), "")
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
	t.Run("overwrite", func(t *testing.T) {
		db, _ := New(fnv.Hasher{})
		db.Set("key1", "Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
		db.Set("key2", "Aenean congue quis nisl ut vulputate. Sed lacus dolor, tempor nec elit sit amet, congue dapibus purus. Pellentesque a lectus vel leo finibus scelerisque.")
		db.Set("key3", "Aliquam blandit mauris mauris, eget bibendum lacus tempus non. Duis orci leo, sagittis sed lorem eu, pulvinar elementum leo.")

		var lo, hi uint32

		hkey := db.hasher.Sum64("key1")
		e := db.setLF(hkey, "Nunc lacinia, purus finibus consectetur ullamcorper, nisi elit laoreet augue, vitae tincidunt tellus velit sit amet arcu.")
		lo, hi = e.decode()
		assertLH(t, lo, hi, 3, 4)

		t9n := "Quisque sit amet viverra ligula. Praesent sagittis, sapien ut rutrum porttitor, dolor ligula accumsan velit, ut lacinia tellus tellus nec tortor."
		hkey = db.hasher.Sum64("key2")
		e = db.setLF(hkey, t9n)
		lo, hi = e.decode()
		assertLH(t, lo, hi, 1, 2)
		assertT9n(t, db, "key2", t9n)
	})
}

func TestPlural(t *testing.T) {
	testPlural := func(db *DB, key, def string, count int, expect string) {
		repl := PlaceholderReplacer{}
		repl.AddKV("!count", strconv.Itoa(count))
		s := db.GetPluralWR(key, def, count, &repl)
		if s != expect {
			t.Errorf("plural mismatch, need %s got %s", expect, s)
		}
	}

	db, _ := New(fnv.Hasher{})
	db.SetPolicy(policy.Locked)
	db.Set("en.user.bag.apples_flag", "You have one apple|You have many apples")
	db.Set("en.user.bag.apples", "You have !count apple|You have !count apples")
	db.Set("en.h3.army_size", "[1,5] Few|[5,10] Several|[10,20] Pack|[20,50] Lots|[50,100] Horde|[100,250] Throng|[250,500] Swarm|[500,1000] Zounds|[1000,*] Legion")
	db.Set("ru.user.bag.apples", "[*,0] У вас проблемы с математикой|{0} У вас нет яблок|{1} У вас !count яблоко|[2,5] У вас !count яблока|[5,21] У вас !count яблок|{21} У вас !count яблоко|[22,25] У вас !count яблока|[25,*] У вас много яблок")
	db.SetPolicy(policy.LockFree)

	t.Run("en.simple[1]", func(t *testing.T) { testPlural(db, "en.user.bag.apples_flag", "", 1, "You have one apple") })
	t.Run("en.simple[2]", func(t *testing.T) { testPlural(db, "en.user.bag.apples_flag", "", 2, "You have many apples") })

	t.Run("en.placeholder[1]", func(t *testing.T) { testPlural(db, "en.user.bag.apples", "", 1, "You have 1 apple") })
	t.Run("en.placeholder[5]", func(t *testing.T) { testPlural(db, "en.user.bag.apples", "", 5, "You have 5 apples") })

	t.Run("en.h3.enemy_size[0]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "N/D", 0, "N/D") })
	t.Run("en.h3.enemy_size[2]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "", 2, "Few") })
	t.Run("en.h3.enemy_size[19]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "", 19, "Pack") })
	t.Run("en.h3.enemy_size[20]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "", 20, "Lots") })
	t.Run("en.h3.enemy_size[333]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "", 333, "Swarm") })
	t.Run("en.h3.enemy_size[999]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "", 999, "Zounds") })
	t.Run("en.h3.enemy_size[1e9]", func(t *testing.T) { testPlural(db, "en.h3.army_size", "", 1e9, "Legion") })

	t.Run("ru.placeholder[-15]", func(t *testing.T) {
		testPlural(db, "ru.user.bag.apples", "", -15, "У вас проблемы с математикой")
	})
	t.Run("ru.placeholder[0]", func(t *testing.T) { testPlural(db, "ru.user.bag.apples", "", 0, "У вас нет яблок") })
	t.Run("ru.placeholder[1]", func(t *testing.T) { testPlural(db, "ru.user.bag.apples", "", 1, "У вас 1 яблоко") })
	t.Run("ru.placeholder[3]", func(t *testing.T) { testPlural(db, "ru.user.bag.apples", "", 3, "У вас 3 яблока") })
	t.Run("ru.placeholder[11]", func(t *testing.T) { testPlural(db, "ru.user.bag.apples", "", 11, "У вас 11 яблок") })
	t.Run("ru.placeholder[21]", func(t *testing.T) { testPlural(db, "ru.user.bag.apples", "", 21, "У вас 21 яблоко") })
	t.Run("ru.placeholder[24]", func(t *testing.T) { testPlural(db, "ru.user.bag.apples", "", 24, "У вас 24 яблока") })
	t.Run("ru.placeholder[999999]", func(t *testing.T) {
		testPlural(db, "ru.user.bag.apples", "", 999999, "У вас много яблок")
	})
}

func BenchmarkIO(b *testing.B) {
	benchIO := func(b *testing.B, entries int64) {
		buf := []byte("en.")
		db, _ := New(fnv.Hasher{})
		db.SetPolicy(policy.Locked)
		for i := int64(0); i < entries; i++ {
			buf = strconv.AppendInt(buf[:3], i, 10)
			db.Set(fastconv.B2S(buf), "Hello there!")
		}
		db.SetPolicy(policy.LockFree)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			x := rand.Int63n(entries)
			buf = strconv.AppendInt(buf[:3], x, 10)
			s := db.Get(fastconv.B2S(buf), "")
			if s != "Hello there!" {
				b.Error("translation mismatch")
			}
		}
	}

	b.Run("1", func(b *testing.B) { benchIO(b, 1) })
	b.Run("10", func(b *testing.B) { benchIO(b, 10) })
	b.Run("100", func(b *testing.B) { benchIO(b, 100) })
	b.Run("1K", func(b *testing.B) { benchIO(b, 1000) })
	b.Run("10K", func(b *testing.B) { benchIO(b, 10000) })
	b.Run("100K", func(b *testing.B) { benchIO(b, 100000) })
}
