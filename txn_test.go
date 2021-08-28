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

	txn := db.Begin()
	for i := 5; i < 15; i++ {
		buf = strconv.AppendInt(buf[:3], int64(i), 10)
		txn.Set(fastconv.B2S(buf), "qwerty")
	}

	if txn.Size() != 10 {
		t.Error("txn size failed, need 10 got", txn.Size())
	}
	if !bytes.Equal(txn.data, []byte("en.5qwertyen.6qwertyen.7qwertyen.8qwertyen.9qwertyen.10qwertyen.11qwertyen.12qwertyen.13qwertyen.14qwerty")) {
		t.Error("txn contents mismatch")
	}

	txn.Commit()

	buf = strconv.AppendInt(buf[:3], 12, 10)
	s := db.Get(fastconv.B2S(buf))
	if s != "qwerty" {
		t.Error("db updated entry mismatch, need qwerty got", s)
	}
}
