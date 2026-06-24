package reg_test

import (
	"testing"

	"github.com/kjkrol/goke/v2/internal/reg"
)

type Position struct{ X, Y float64 }
type Velocity struct{ VX, VY float64 }
type Tag struct{}

func newRegistry(t *testing.T) *reg.Registry {
	t.Helper()
	var r reg.Registry
	r.Init(reg.DefaultConfig())
	return &r
}
