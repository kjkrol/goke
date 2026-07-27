package bench_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke/v2"
)

// Benchmark_Editor_Del measures the per-entity cost of a structural remove via
// Editor.Update. Entities are pre-populated with 10 components at setup; each
// sub-benchmark deletes comps=N of them (timed) then adds them back outside
// the timer so every loop iteration starts from the same 10-component state. Each
// sub-benchmark runs in two orders: sorted iterates entities in creation
// order, random reshuffles them before every iteration (outside the timer).
func Benchmark_Editor_Del(b *testing.B) {
	ecs := setupECS()

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=1/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=2/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=3/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=4/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=5/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=6/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=7/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=8/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]), new(goke.Comp[T08])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=9/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08](), goke.Del[T09]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]), new(goke.Comp[T08]), new(goke.Comp[T09])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=10/comps=10/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populate(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))

			// Anchor entities with an extra untouched component so deleting all 10
			// tracked components never leaves zero components: Editor.Update unlinks
			// (destroys) an entity that would end up with none, which would silently
			// no-op the untimed addEd restore below and corrupt every iteration after
			// the first.
			anchorEd := ecs.NewEditorBuilder(new(goke.Comp[Base])).Build()
			for _, e := range entities {
				anchorEd.Update(e)
			}

			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08](), goke.Del[T09](), goke.Del[T10]()).Build()
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Pos]), new(goke.Comp[Vel]), new(goke.Comp[Acc]), new(goke.Comp[T04]), new(goke.Comp[T05]), new(goke.Comp[T06]), new(goke.Comp[T07]), new(goke.Comp[T08]), new(goke.Comp[T09]), new(goke.Comp[T10])).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
}
