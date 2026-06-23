package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

func populateBase(ecs *goke.ECS, count int) []uid.UID64 {
	factory := ecs.CreateFactory(goke.Add(new(goke.Col[Base])))
	var ids []uid.UID64
	factory.Create(count)
	for factory.Next() {
		ids = append(ids, factory.IDs...)
	}
	return ids
}

// Benchmark_Editor_Add measures the per-entity cost of a structural add via
// Editor.Update. Entities start with only a Base anchor; each iteration adds N
// components (timed) then resets back to Base outside the timer.
func Benchmark_Editor_Add(b *testing.B) {
	ecs := setupECS()

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=1", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		addEd := ecs.CreateEditor(goke.Add(&c1))
		delEd := ecs.CreateEditor(goke.Del[Pos]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=2", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=3", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=4", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3), goke.Add(&c4))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=5", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3), goke.Add(&c4), goke.Add(&c5))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
					c5.At(cur).V = 5
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=6", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3),
			goke.Add(&c4), goke.Add(&c5), goke.Add(&c6))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
					c5.At(cur).V = 5
					c6.At(cur).V = 6
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=7", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3),
			goke.Add(&c4), goke.Add(&c5), goke.Add(&c6), goke.Add(&c7))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
					c5.At(cur).V = 5
					c6.At(cur).V = 6
					c7.At(cur).V = 7
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=8", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3),
			goke.Add(&c4), goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7), goke.Add(&c8))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
					c5.At(cur).V = 5
					c6.At(cur).V = 6
					c7.At(cur).V = 7
					c8.At(cur).V = 8
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=9", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		var c9 goke.Col[T09]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3),
			goke.Add(&c4), goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7), goke.Add(&c8), goke.Add(&c9))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08](), goke.Del[T09]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
					c5.At(cur).V = 5
					c6.At(cur).V = 6
					c7.At(cur).V = 7
					c8.At(cur).V = 8
					c9.At(cur).V = 9
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=1/add=10", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		var c1 goke.Col[Pos]
		var c2 goke.Col[Vel]
		var c3 goke.Col[Acc]
		var c4 goke.Col[T04]
		var c5 goke.Col[T05]
		var c6 goke.Col[T06]
		var c7 goke.Col[T07]
		var c8 goke.Col[T08]
		var c9 goke.Col[T09]
		var c10 goke.Col[T10]
		addEd := ecs.CreateEditor(goke.Add(&c1), goke.Add(&c2), goke.Add(&c3),
			goke.Add(&c4), goke.Add(&c5), goke.Add(&c6),
			goke.Add(&c7), goke.Add(&c8), goke.Add(&c9), goke.Add(&c10))
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08](), goke.Del[T09](), goke.Del[T10]())
		cur := &addEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					addEd.Update(e)
					c1.At(cur).X = 1
					c2.At(cur).X = 2
					c3.At(cur).X = 3
					c4.At(cur).V = 4
					c5.At(cur).V = 5
					c6.At(cur).V = 6
					c7.At(cur).V = 7
					c8.At(cur).V = 8
					c9.At(cur).V = 9
					c10.At(cur).V = 10
				}
				b.StopTimer()
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})
}
