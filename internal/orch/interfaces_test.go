package orch

import (
	"testing"
	"time"
)

func TestRunnableFunc_Update(t *testing.T) {
	var gotCB *CmdBuf
	var gotD time.Duration

	fn := RunnableFunc(func(cb *CmdBuf, d time.Duration) {
		gotCB = cb
		gotD = d
	})

	var r Runnable = &fn
	cb := &CmdBuf{}
	r.Update(cb, 5*time.Millisecond)

	if gotCB != cb {
		t.Error("expected Update to pass through the given *CmdBuf")
	}
	if gotD != 5*time.Millisecond {
		t.Errorf("expected Update to pass through the given duration, got %v", gotD)
	}
}
