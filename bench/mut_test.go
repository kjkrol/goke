package bench_test

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// writer is a system's fields as Benchmark_Mut writes through them: Pos by Comp in byComp, by
// Mut in byMut.
type writer struct {
	byComp, byMut *goke.Query
	pos           goke.Comp[Pos]
	mut           goke.Mut[Pos]
	vel           goke.Comp[Vel]
}

func newWriter(si *goke.SysInit) *writer {
	w := &writer{}
	w.byComp = si.NewQueryBuilder(&w.pos, &w.vel).Build()
	w.byMut = si.NewQueryBuilder(&w.mut, &w.vel).Build()
	return w
}

func (w *writer) allComp() {
	cur := w.byComp.Cursor()
	for w.byComp.All(); w.byComp.Next(); {
		p := w.pos.Slice(cur)
		v := w.vel.Slice(cur)
		for i, e := range cur.IDs {
			_ = e
			p[i].X += v[i].X
		}
	}
}

func (w *writer) allMut() {
	cur := w.byMut.Cursor()
	for w.byMut.All(); w.byMut.Next(); {
		p := w.mut.Slice(cur)
		v := w.vel.Slice(cur)
		for i, e := range cur.IDs {
			_ = e
			p[i].X += v[i].X
		}
	}
}

// allMutLast takes the Mut slice after the Comp one, which then keeps its bounds checks.
func (w *writer) allMutLast() {
	cur := w.byMut.Cursor()
	for w.byMut.All(); w.byMut.Next(); {
		v := w.vel.Slice(cur)
		p := w.mut.Slice(cur)
		for i, e := range cur.IDs {
			_ = e
			p[i].X += v[i].X
		}
	}
}

func (w *writer) seekComp(ids []uid.UID64) {
	cur := w.byComp.Cursor()
	for _, e := range ids {
		if w.byComp.Seek(e) {
			w.pos.At(cur).X += w.vel.At(cur).X
		}
	}
}

func (w *writer) seekMut(ids []uid.UID64) {
	cur := w.byMut.Cursor()
	for _, e := range ids {
		if w.byMut.Seek(e) {
			w.mut.At(cur).X += w.vel.At(cur).X
		}
	}
}

func (w *writer) seekHComp(ids []uid.UID64) {
	cur := w.byComp.Cursor()
	w.byComp.Seek(ids[0])
	for _, e := range ids {
		if w.byComp.SeekH(e) {
			w.pos.At(cur).X += w.vel.At(cur).X
		}
	}
}

func (w *writer) seekHMut(ids []uid.UID64) {
	cur := w.byMut.Cursor()
	w.byMut.Seek(ids[0])
	for _, e := range ids {
		if w.byMut.SeekH(e) {
			w.mut.At(cur).X += w.vel.At(cur).X
		}
	}
}

// Benchmark_Mut compares a system writing a column through Comp with one writing it through
// Mut, which also counts a change of each chunk it reaches: All over 1,024 and 65,536 entities
// (3 and 153 chunks), Seek and SeekH over 512 of 1,024 in random order. Mut-last takes the Mut
// slice after a Comp one. sync is an empty Sync over 65,536 entities, without and with
// WithMutCheck.
func Benchmark_Mut(b *testing.B) {
	for _, pop := range []int{entitiesNumber, 64 * entitiesNumber} {
		var w *writer
		setupECS().Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
			populate(si, pop)
			w = newWriter(si)
		}})
		for _, c := range []struct {
			name string
			walk func()
		}{{"Comp", w.allComp}, {"Mut", w.allMut}, {"Mut-last", w.allMutLast}} {
			b.Run(fmt.Sprintf("all/pop=%d/%s", pop, c.name), func(b *testing.B) {
				measurePerEntity(b, pop, func() {
					for b.Loop() {
						c.walk()
					}
				})
			})
		}
	}

	var w *writer
	var entities []uid.UID64
	setupECS().Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		entities = populate(si, entitiesNumber)
		w = newWriter(si)
	}})
	ids := append([]uid.UID64(nil), entities...)
	shuffleIDs(rand.New(rand.NewPCG(42, 1337)), ids)
	ids = ids[:filterSubsetSize]
	for _, c := range []struct {
		name string
		seek func([]uid.UID64)
	}{
		{"seek/%s/Comp", w.seekComp}, {"seek/%s/Mut", w.seekMut},
		{"seekh/%s/Comp", w.seekHComp}, {"seekh/%s/Mut", w.seekHMut},
	} {
		name := fmt.Sprintf(c.name, fmt.Sprintf("pop=%d/subset=%d", entitiesNumber, filterSubsetSize))
		b.Run(name, func(b *testing.B) {
			measurePerEntity(b, filterSubsetSize, func() {
				for b.Loop() {
					c.seek(ids)
				}
			})
		})
	}

	for _, opts := range [][]goke.ECSOption{nil, {goke.WithMutCheck()}} {
		ecs := goke.New(opts...)
		pop := 64 * entitiesNumber
		ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) { populate(si, pop) }})
		ecs.SetPlan(func(ctx goke.RunCtx, _ time.Duration) { _ = ctx.Sync() })
		b.Run(fmt.Sprintf("sync/pop=%d/check=%v", pop, opts != nil), func(b *testing.B) {
			measurePerEntity(b, pop, func() {
				for b.Loop() {
					ecs.Tick(0)
				}
			})
		})
	}
}
