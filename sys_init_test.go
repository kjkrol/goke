package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v3"
)

// TestSysInit_RegComp confirms SysInit.RegComp[T] registers the same
// component as ECS.RegComp[T] — the documented way to register lazily from
// within Init.
func TestSysInit_RegComp(t *testing.T) {
	ecs := goke.New()

	var compID goke.CompID
	ecs.Setup(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			compID = si.RegComp[Position]()
		},
	})

	if compID != ecs.RegComp[Position]() {
		t.Errorf("expected RegComp via si.RegComp[T] to resolve the same CompID as ecs.RegComp[T] directly")
	}
}
