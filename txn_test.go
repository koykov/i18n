package i18n

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/koykov/fastconv"
	"github.com/koykov/hash/fnv"
)

func TestTXN(t *testing.T) {
	buf := []byte("en.")
	db, _ := New(fnv.Hasher{})
	for i := 0; i < 10; i++ {
		buf = strconv.AppendInt(buf[:3], int64(i), 10)
		db.Set(fastconv.B2S(buf), "foo bar")
	}

	db.BeginTXN()
	for i := 5; i < 15; i++ {
		translation := "qwerty"
		if i == 9 {
			translation = "foo bar"
		}
		buf = strconv.AppendInt(buf[:3], int64(i), 10)
		db.Set(fastconv.B2S(buf), translation)
	}

	txn := (*txn)(db.txn)
	if txn.size() != 9 {
		t.Error("txn size failed, need 9 got", txn.size())
	}
	if !bytes.Equal(txn.data, []byte("qwertyqwertyqwertyqwertyqwertyqwertyqwertyqwertyqwerty")) {
		t.Error("txn contents mismatch")
	}

	db.Commit()

	buf = strconv.AppendInt(buf[:3], 12, 10)
	s := db.Get(fastconv.B2S(buf))
	if s != "qwerty" {
		t.Error("db updated entry mismatch, need qwerty got", s)
	}
}
