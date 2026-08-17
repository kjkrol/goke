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
			compID = goke.RegComp[Position](si.ECS())
		},
	})

	if compID != goke.RegComp[Position](ecs) {
		t.Errorf("expected RegComp via si.ECS() to resolve the same CompID as RegComp(ecs) directly")
	}
}
