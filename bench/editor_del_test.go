package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
)

// Benchmark_Editor_Del measures the per-entity cost of a structural remove via
// Editor.Update. Entities are pre-populated with 10 components at setup; each
// sub-benchmark deletes N of them (timed) then adds them back outside the timer
// so every loop iteration starts from the same 10-component state.
func Benchmark_Editor_Del(b *testing.B) {
	ecs := setupECS()

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=1", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=2", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=3", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=4", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])),
			goke.Add(new(goke.Col[Acc])), goke.Add(new(goke.Col[T04])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=5", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=6", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=7", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=8", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])), goke.Add(new(goke.Col[T08])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=9", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)
		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08](), goke.Del[T09]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])), goke.Add(new(goke.Col[T08])), goke.Add(new(goke.Col[T09])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/del=10", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populate(ecs, entitiesNumber)

		// Anchor entities with an extra untouched component so deleting all 10
		// tracked components never leaves zero components: Editor.Update unlinks
		// (destroys) an entity that would end up with none, which would silently
		// no-op the untimed addEd restore below and corrupt every iteration after
		// the first.
		anchorEd := ecs.CreateEditor(goke.Add(new(goke.Col[Base])))
		for _, e := range entities {
			anchorEd.Update(e)
		}

		delEd := ecs.CreateEditor(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08](), goke.Del[T09](), goke.Del[T10]())
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])), goke.Add(new(goke.Col[T08])), goke.Add(new(goke.Col[T09])),
			goke.Add(new(goke.Col[T10])))
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					delEd.Update(e)
				}
				b.StopTimer()
				for _, e := range entities {
					addEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})
}
