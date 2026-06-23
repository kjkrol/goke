package ent_test

import (
	"testing"

	"github.com/kjkrol/goke/internal/ent"
)

func TestDefaultConfig(t *testing.T) {
	cfg := ent.DefaultConfig()

	if cfg.Cap != 1000 {
		t.Errorf("expected Cap 1000, got %d", cfg.Cap)
	}
	if cfg.FreeCap != 1024 {
		t.Errorf("expected FreeCap 1024, got %d", cfg.FreeCap)
	}
}
