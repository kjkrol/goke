package orch

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/kjkrol/goke/v3/internal/comp"
	"github.com/kjkrol/uid"
)

func TestCmdBuf_Reserve(t *testing.T) {
	type tag struct{}
	var defs comp.DefIndex
	defs.Init()
	dataID := defs.Intern(reflect.TypeFor[mockCompA]()).ID
	stringID := defs.Intern(reflect.TypeFor[mockCompB]()).ID
	tagID := defs.Intern(reflect.TypeFor[tag]()).ID
	cb := NewCmdBuf()
	cb.SetDefs(&defs)

	p := cb.Reserve(uid.UID64(1), dataID)
	*(*mockCompA)(p) = mockCompA{Val: 9}
	q := cb.Reserve(uid.UID64(2), stringID)
	*(*mockCompB)(q) = mockCompB{Msg: "staged"}
	if r := cb.Reserve(uid.UID64(3), tagID); r != nil {
		t.Errorf("expected a nil slot for a tag, got %v", r)
	}

	if len(cb.cmds) != 3 {
		t.Fatalf("expected 3 queued commands, got %d", len(cb.cmds))
	}
	data := cb.cmds[0]
	if data.cType != cmdAssignComp || data.entityID != 1 || data.compID != dataID {
		t.Errorf("unexpected command %+v", data)
	}
	if data.dataPtr != p || data.size != unsafe.Sizeof(mockCompA{}) {
		t.Errorf("expected the staged slot %v of %d bytes, got %v of %d", p, unsafe.Sizeof(mockCompA{}), data.dataPtr, data.size)
	}
	if (*mockCompA)(data.dataPtr).Val != 9 {
		t.Errorf("the value written into the slot is not staged")
	}
	if (*mockCompB)(cb.cmds[1].dataPtr).Msg != "staged" {
		t.Errorf("the string value written into the slot is not staged")
	}
	if cb.cmds[2].dataPtr != nil || cb.cmds[2].size != 0 {
		t.Errorf("expected no payload for a tag, got %+v", cb.cmds[2])
	}
}
