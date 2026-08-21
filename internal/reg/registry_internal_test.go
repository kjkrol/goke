package reg

import "testing"

// TestRegistry_Reset_PanicsWhileSaving exercises the saving guard directly
// (package reg, not reg_test) — under normal, single-goroutine usage saving
// is never true when Reset is called (Save doesn't call back into user
// code while it's set), so this invariant has no reachable trigger via the
// public API alone.
func TestRegistry_Reset_PanicsWhileSaving(t *testing.T) {
	var r Registry
	r.Init(DefaultConfig())
	r.saving = true

	defer func() {
		if recover() == nil {
			t.Error("expected Reset to panic while a Save is in progress")
		}
	}()
	r.Reset()
}
