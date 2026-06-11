package system

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
)

type System interface {
	Update(core.ReadOnlyRegistry, *SystemCommandBuffer, time.Duration)
}
