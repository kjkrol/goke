package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

func populateBase(ecs *goke.ECS, count int) []uid.UID64 {
	factory := ecs.NewFactory(new(goke.Comp[Base]))
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
		var c1 goke.Comp[Pos]
		addEd := ecs.NewEditorBuilder(&c1).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		addEd := ecs.NewEditorBuilder(&c1, &c2).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4, &c5).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4, &c5, &c6).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4, &c5, &c6, &c7).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		var c8 goke.Comp[T08]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		var c8 goke.Comp[T08]
		var c9 goke.Comp[T09]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08](), goke.Del[T09]()).Build()
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
		var c1 goke.Comp[Pos]
		var c2 goke.Comp[Vel]
		var c3 goke.Comp[Acc]
		var c4 goke.Comp[T04]
		var c5 goke.Comp[T05]
		var c6 goke.Comp[T06]
		var c7 goke.Comp[T07]
		var c8 goke.Comp[T08]
		var c9 goke.Comp[T09]
		var c10 goke.Comp[T10]
		addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10).Build()
		delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08](), goke.Del[T09](), goke.Del[T10]()).Build()
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
