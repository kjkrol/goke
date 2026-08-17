package control

import (
	"time"

	"github.com/kjkrol/goke/v2"
)

// DefaultController captures input each frame and, once an EventHandler is
// registered via SetHandler, runs it every tick.
type DefaultController struct {
	adapter InputAdapter
	events  *InputEvents
	handler EventHandler
}

func NewDefaultController(adapter InputAdapter, events *InputEvents) *DefaultController {
	return &DefaultController{adapter: adapter, events: events}
}

func (c *DefaultController) SetHandler(handler EventHandler) { c.handler = handler }

func (c *DefaultController) Capture(e *InputEvents) { c.adapter.Capture(e) }

func (c *DefaultController) Init(*goke.SysInit) {}

func (c *DefaultController) Update(_ *goke.CmdBuf, _ time.Duration) {
	if c.handler != nil {
		c.handler.HandleEvents(c.events)
	}
}
