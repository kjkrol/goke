package exec

import (
	"time"

	"github.com/kjkrol/goke/internal/core"
)

type Runnable interface {
	Update(core.ComponentReader, *CommandBuf, time.Duration)
}
