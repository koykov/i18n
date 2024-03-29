package i18n

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/koykov/byteconv"
	"github.com/koykov/hash/fnv"
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

func assertT9nPlural(t *testing.T, db *DB, key, expect string, count int) {
	t9n := db.GetPlural(key, "", count)
	if t9n != expect {
		t.Errorf("translation mismatch, need %s, got %s", expect, t9n)
	}
}

func TestIO(t *testing.T) {
	testIO := func(t *testing.T, entries int64) {
		buf := []byte("en.")
		db, _ := New(fnv.Hasher{})
		for i := int64(0); i < entries; i++ {
			buf = strconv.AppendInt(buf[:3], i, 10)
			_ = db.Set(byteconv.B2S(buf), "Hello there!")
		}

		i := rand.Int63n(entries)
		buf = strconv.AppendInt(buf[:3], i, 10)
		s := db.Get(byteconv.B2S(buf), "")
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
		_ = db.Set("key1", "Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
		_ = db.Set("key2", "Aenean congue quis nisl ut vulputate. Sed lacus dolor, tempor nec elit sit amet, congue dapibus purus. Pellentesque a lectus vel leo finibus scelerisque.")
		_ = db.Set("key3", "Aliquam blandit mauris mauris, eget bibendum lacus tempus non. Duis orci leo, sagittis sed lorem eu, pulvinar elementum leo.")

		var lo, hi uint32

		hkey := db.hasher.Sum64("key1")
		e := db.setLF(hkey, "Nunc lacinia, purus finibus consectetur ullamcorper, nisi elit laoreet augue, vitae tincidunt tellus velit sit amet arcu.")
		lo, hi = e.Decode()
		assertLH(t, lo, hi, 3, 4)

		t9n := "Quisque sit amet viverra ligula. Praesent sagittis, sapien ut rutrum porttitor, dolor ligula accumsan velit, ut lacinia tellus tellus nec tortor."
		hkey = db.hasher.Sum64("key2")
		e = db.setLF(hkey, t9n)
		lo, hi = e.Decode()
		assertLH(t, lo, hi, 1, 2)
		assertT9n(t, db, "key2", t9n)
	})
	t.Run("overwrite_plural", func(t *testing.T) {
		db, _ := New(fnv.Hasher{})
		_ = db.Set("key1", "There is one apple|There are many apples")
		_ = db.Set("key2", "{0} There are none|[1,19] There are some|[20,*] There are many")
		_ = db.Set("key3", "{1} :value minute ago|[2,*] :value minutes ago")

		var lo, hi uint32

		hkey := db.hasher.Sum64("key1")
		e := db.setLF(hkey, "{0} There are none|{1} There is one|[2,*] There are :count")
		lo, hi = e.Decode()
		assertLH(t, lo, hi, 7, 10)
		assertT9nPlural(t, db, "key1", "There is one", 1)

		t9n := "{0} There are none|{1} There is one|[2,*] There are :count"
		hkey = db.hasher.Sum64("key2")
		e = db.setLF(hkey, t9n)
		lo, hi = e.Decode()
		assertLH(t, lo, hi, 2, 5)
		assertT9nPlural(t, db, "key2", "There is one", 1)
		assertT9nPlural(t, db, "key2", "There are :count", 10)
	})
}

func TestPlural(t *testing.T) {
	testPlural := func(t *testing.T, db *DB, key, def string, count int, expect string) {
		repl := PlaceholderReplacer{}
		repl.AddKV("!count", strconv.Itoa(count))
		s := db.GetPluralWR(key, def, count, &repl)
		if s != expect {
			t.Errorf("plural mismatch, need %s got %s", expect, s)
		}
	}

	db, _ := New(fnv.Hasher{})
	_ = db.Set("en.user.bag.apples_flag", "You have one apple|You have many apples")
	_ = db.Set("en.user.bag.apples", "You have !count apple|You have !count apples")
	_ = db.Set("en.h3.army_size", "[1,5] Few|[5,10] Several|[10,20] Pack|[20,50] Lots|[50,100] Horde|[100,250] Throng|[250,500] Swarm|[500,1000] Zounds|[1000,*] Legion")
	_ = db.Set("ru.user.bag.apples", "[*,0] У вас проблемы с математикой|{0} У вас нет яблок|{1} У вас !count яблоко|[2,5] У вас !count яблока|[5,21] У вас !count яблок|{21} У вас !count яблоко|[22,25] У вас !count яблока|[25,*] У вас много яблок")

	t.Run("en.simple[1]", func(t *testing.T) { testPlural(t, db, "en.user.bag.apples_flag", "", 1, "You have one apple") })
	t.Run("en.simple[2]", func(t *testing.T) { testPlural(t, db, "en.user.bag.apples_flag", "", 2, "You have many apples") })

	t.Run("en.placeholder[1]", func(t *testing.T) { testPlural(t, db, "en.user.bag.apples", "", 1, "You have 1 apple") })
	t.Run("en.placeholder[5]", func(t *testing.T) { testPlural(t, db, "en.user.bag.apples", "", 5, "You have 5 apples") })

	t.Run("en.h3.enemy_size[0]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "N/D", 0, "N/D") })
	t.Run("en.h3.enemy_size[2]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "", 2, "Few") })
	t.Run("en.h3.enemy_size[19]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "", 19, "Pack") })
	t.Run("en.h3.enemy_size[20]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "", 20, "Lots") })
	t.Run("en.h3.enemy_size[333]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "", 333, "Swarm") })
	t.Run("en.h3.enemy_size[999]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "", 999, "Zounds") })
	t.Run("en.h3.enemy_size[1e9]", func(t *testing.T) { testPlural(t, db, "en.h3.army_size", "", 1e9, "Legion") })

	t.Run("ru.placeholder[-15]", func(t *testing.T) {
		testPlural(t, db, "ru.user.bag.apples", "", -15, "У вас проблемы с математикой")
	})
	t.Run("ru.placeholder[0]", func(t *testing.T) { testPlural(t, db, "ru.user.bag.apples", "", 0, "У вас нет яблок") })
	t.Run("ru.placeholder[1]", func(t *testing.T) { testPlural(t, db, "ru.user.bag.apples", "", 1, "У вас 1 яблоко") })
	t.Run("ru.placeholder[3]", func(t *testing.T) { testPlural(t, db, "ru.user.bag.apples", "", 3, "У вас 3 яблока") })
	t.Run("ru.placeholder[11]", func(t *testing.T) { testPlural(t, db, "ru.user.bag.apples", "", 11, "У вас 11 яблок") })
	t.Run("ru.placeholder[21]", func(t *testing.T) { testPlural(t, db, "ru.user.bag.apples", "", 21, "У вас 21 яблоко") })
	t.Run("ru.placeholder[24]", func(t *testing.T) { testPlural(t, db, "ru.user.bag.apples", "", 24, "У вас 24 яблока") })
	t.Run("ru.placeholder[999999]", func(t *testing.T) {
		testPlural(t, db, "ru.user.bag.apples", "", 999999, "У вас много яблок")
	})
}

func BenchmarkIO(b *testing.B) {
	benchIO := func(b *testing.B, entries int64) {
		buf := []byte("en.")
		db, _ := New(fnv.Hasher{})
		for i := int64(0); i < entries; i++ {
			buf = strconv.AppendInt(buf[:3], i, 10)
			_ = db.Set(byteconv.B2S(buf), "Hello there!")
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			x := rand.Int63n(entries)
			buf = strconv.AppendInt(buf[:3], x, 10)
			s := db.Get(byteconv.B2S(buf), "")
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

func BenchmarkPlural(b *testing.B) {
	benchPlural := func(b *testing.B, db *DB, key, def string, count int, expect string) {
		repl := PlaceholderReplacer{}
		sc := strconv.Itoa(count)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			repl.Reset()
			repl.AddKV("!count", sc)
			s := db.GetPluralWR(key, def, count, &repl)
			if s != expect {
				b.Errorf("plural mismatch, need %s got %s", expect, s)
			}
		}
	}

	db, _ := New(fnv.Hasher{})
	_ = db.Set("en.user.bag.apples_flag", "You have one apple|You have many apples")
	_ = db.Set("en.user.bag.apples", "You have !count apple|You have !count apples")
	_ = db.Set("en.h3.army_size", "[1,5] Few|[5,10] Several|[10,20] Pack|[20,50] Lots|[50,100] Horde|[100,250] Throng|[250,500] Swarm|[500,1000] Zounds|[1000,*] Legion")
	_ = db.Set("ru.user.bag.apples", "[*,0] У вас проблемы с математикой|{0} У вас нет яблок|{1} У вас !count яблоко|[2,5] У вас !count яблока|[5,21] У вас !count яблок|{21} У вас !count яблоко|[22,25] У вас !count яблока|[25,*] У вас много яблок")

	b.Run("en.simple[1]", func(b *testing.B) { benchPlural(b, db, "en.user.bag.apples_flag", "", 1, "You have one apple") })
	b.Run("en.simple[2]", func(b *testing.B) { benchPlural(b, db, "en.user.bag.apples_flag", "", 2, "You have many apples") })

	b.Run("en.placeholder[1]", func(b *testing.B) { benchPlural(b, db, "en.user.bag.apples", "", 1, "You have 1 apple") })
	b.Run("en.placeholder[5]", func(b *testing.B) { benchPlural(b, db, "en.user.bag.apples", "", 5, "You have 5 apples") })

	b.Run("en.h3.enemy_size[0]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "N/D", 0, "N/D") })
	b.Run("en.h3.enemy_size[2]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "", 2, "Few") })
	b.Run("en.h3.enemy_size[19]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "", 19, "Pack") })
	b.Run("en.h3.enemy_size[20]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "", 20, "Lots") })
	b.Run("en.h3.enemy_size[333]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "", 333, "Swarm") })
	b.Run("en.h3.enemy_size[999]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "", 999, "Zounds") })
	b.Run("en.h3.enemy_size[1e9]", func(b *testing.B) { benchPlural(b, db, "en.h3.army_size", "", 1e9, "Legion") })

	b.Run("ru.placeholder[-15]", func(b *testing.B) {
		benchPlural(b, db, "ru.user.bag.apples", "", -15, "У вас проблемы с математикой")
	})
	b.Run("ru.placeholder[0]", func(b *testing.B) { benchPlural(b, db, "ru.user.bag.apples", "", 0, "У вас нет яблок") })
	b.Run("ru.placeholder[1]", func(b *testing.B) { benchPlural(b, db, "ru.user.bag.apples", "", 1, "У вас 1 яблоко") })
	b.Run("ru.placeholder[3]", func(b *testing.B) { benchPlural(b, db, "ru.user.bag.apples", "", 3, "У вас 3 яблока") })
	b.Run("ru.placeholder[11]", func(b *testing.B) { benchPlural(b, db, "ru.user.bag.apples", "", 11, "У вас 11 яблок") })
	b.Run("ru.placeholder[21]", func(b *testing.B) { benchPlural(b, db, "ru.user.bag.apples", "", 21, "У вас 21 яблоко") })
	b.Run("ru.placeholder[24]", func(b *testing.B) { benchPlural(b, db, "ru.user.bag.apples", "", 24, "У вас 24 яблока") })
	b.Run("ru.placeholder[999999]", func(b *testing.B) {
		benchPlural(b, db, "ru.user.bag.apples", "", 999999, "У вас много яблок")
	})
}
