package sqlstore

import (
	"github.com/gasparian/money-transfers-api/internal/app/store"
	"os"
	"testing"
)

func TestSqlStore(t *testing.T) {
	dbPath := "/tmp/tets.db"
	s, err := New(dbPath, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	defer os.RemoveAll(dbPath)

	store.TestStore(s, t)
	store.TestStoreConcurrentTransfer(s, t)
}
