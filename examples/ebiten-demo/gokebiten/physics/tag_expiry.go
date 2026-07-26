package physics

import (
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

var _ goke.System = (*TagExpirySystem[struct{}])(nil)

// TagExpirySystem removes tag component T, batched per chunk via a
// Migrator, from any entity once its expiry (given by expiresAt) has
// passed. A zero expiresAt is treated as not-yet-initialized and never
// expires — whatever adds T is expected to set it shortly after.
type TagExpirySystem[T any] struct {
	expiresAt func(*T) time.Time
	query     *goke.Query
	tag       goke.Comp[T]
	remove    *goke.Migrator
	matched   []uid.UID64
}

func NewTagExpirySystem[T any](expiresAt func(*T) time.Time) *TagExpirySystem[T] {
	return &TagExpirySystem[T]{expiresAt: expiresAt}
}

func (s *TagExpirySystem[T]) Init(ecs *goke.ECS) {
	s.query = ecs.NewQueryBuilder(&s.tag).Build()
	s.remove = ecs.NewMigratorBuilder().Delete(goke.Del[T]()).Build()
}

func (s *TagExpirySystem[T]) Update(cb *goke.CmdBuf, _ time.Duration) {
	now := time.Now()
	s.query.All()
	for s.query.Next() {
		cursor := s.query.Cursor()
		tags := s.tag.Slice(cursor)
		s.matched = s.matched[:0]
		for i, id := range cursor.IDs {
			exp := s.expiresAt(&tags[i])
			if exp.IsZero() || !now.After(exp) {
				continue
			}
			s.matched = append(s.matched, id)
		}
		if len(s.matched) > 0 {
			snap := s.query.ChunkSnapshot()
			goke.CmdBufMassMigrate(cb, s.remove, snap, s.matched)
		}
	}
}
