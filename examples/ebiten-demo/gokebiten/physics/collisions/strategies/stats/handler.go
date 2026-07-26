package stats

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/physics/collisions"
	"github.com/kjkrol/goke/v2"
)

var _ collisions.CollisionHandler = (*Handler)(nil)

// Handler is a ready-made collision reaction: a shared hit counter.
// Independent of any physics response — compose it with or without
// strategies/elastic.
type Handler struct {
	stats *Stats
}

func NewHandler(stats *Stats) *Handler {
	return &Handler{stats: stats}
}

func (h *Handler) OnCollision(_ *goke.CmdBuf, _ collisions.CollisionEvent) {
	h.stats.Counter++
}
