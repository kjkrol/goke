package debug

import (
	"fmt"
	"io"
	"os"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions"
	"github.com/kjkrol/goke/v2"
)

var _ collisions.CollisionHandler = (*Handler)(nil)

// Formatter renders one CollisionEvent as a log line.
type Formatter func(e collisions.CollisionEvent) string

func defaultFormat(e collisions.CollisionEvent) string {
	return fmt.Sprintf("collision: %v <-> %v", e.EntityA, e.EntityB)
}

// Handler is a ready-made collision reaction: logs each newly-resolved
// contact. Zero-config prints entity IDs to stdout; use WithWriter/WithFormat
// for more detail (positions, velocities, penetration are all on the event).
type Handler struct {
	w      io.Writer
	format Formatter
}

type Option func(*Handler)

func WithWriter(w io.Writer) Option { return func(h *Handler) { h.w = w } }
func WithFormat(f Formatter) Option { return func(h *Handler) { h.format = f } }

func NewHandler(opts ...Option) *Handler {
	h := &Handler{w: os.Stdout, format: defaultFormat}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) OnCollision(_ *goke.CmdBuf, e collisions.CollisionEvent) {
	fmt.Fprintln(h.w, h.format(e))
}
