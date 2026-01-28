package ecs

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
)

// ---------------- Systems ------------------------------- //

type ReadOnlyRegistry = core.ReadOnlyRegistry
type SystemCommandBuffer = core.SystemCommandBuffer

type System interface {
	Update(reg ReadOnlyRegistry, cb *SystemCommandBuffer, d time.Duration)
	Init(*Engine)
}

type SystemFunc func(reg ReadOnlyRegistry, cb *SystemCommandBuffer, d time.Duration)

type functionalSystem struct {
	updateFn SystemFunc
}

func (f *functionalSystem) Init(reg *Engine) {}

func (f *functionalSystem) Update(reg ReadOnlyRegistry, cb *SystemCommandBuffer, d time.Duration) {
	f.updateFn(reg, cb, d)
}

var _ System = (*functionalSystem)(nil)
