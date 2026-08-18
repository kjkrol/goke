package goke_test

import (
	"testing"

	"github.com/kjkrol/goke/v3"
)

// TestSysInit_ECS confirms SysInit.ECS() returns a usable *ECS — the
// documented escape hatch for calling RegComp lazily from within Init.
func TestSysInit_ECS(t *testing.T) {
	ecs := goke.New()

	var compID goke.CompID
	ecs.Setup(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			compID = si.ECS().RegComp[Position]()
		},
	})

	if compID != ecs.RegComp[Position]() {
		t.Errorf("expected RegComp via si.ECS() to resolve the same CompID as ecs.RegComp() directly")
	}
}
