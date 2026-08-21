package reg_test

import (
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/reg"
)

func TestLoadComp_Register_SizeMismatch(t *testing.T) {
	r := newRegistry(t)
	tok := reg.LoadComp[Position]()

	wantSize := uint32(unsafe.Sizeof(Position{})) + 1
	if err := tok.Register(r, &wantSize); err == nil {
		t.Error("expected an error on component size mismatch")
	}
}

type providesPosition struct{}

func (providesPosition) LoadComps() []reg.CompToken {
	return []reg.CompToken{reg.LoadComp[Position]()}
}

type doesNotProvide struct{}

func TestProvidedComps_SkipsNonProviders(t *testing.T) {
	tokens := reg.ProvidedComps(providesPosition{}, doesNotProvide{}, providesPosition{})
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens from the two CompProvider values, got %d", len(tokens))
	}
	for _, tok := range tokens {
		if tok.Name == "" {
			t.Error("expected a non-empty token name")
		}
	}
}
