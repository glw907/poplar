package mailjmap

import (
	"testing"

	"github.com/glw907/poplar/internal/config"
)

func TestNew_AccountName(t *testing.T) {
	b := New(config.AccountConfig{Name: "alice@example.com"})
	if got := b.AccountName(); got != "alice@example.com" {
		t.Errorf("AccountName = %q, want %q", got, "alice@example.com")
	}
}
