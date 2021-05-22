package kvstore

import (
	"github.com/gasparian/money-transfers-api/internal/app/store"
	"testing"
)

func TestKVStore(t *testing.T) {
	s := New()
	store.TestStore(s, t)
	store.TestStoreConcurrentTransfer(s, t)
}
