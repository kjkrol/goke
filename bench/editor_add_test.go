package bench_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/kjkrol/goke/v2"
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
// Editor.Update. Entities start with only a Base anchor; each iteration adds
// the payload (timed) then resets back to Base outside the timer. The payload
// is either comps=N data components or tags=N zero-size tags. Each
// sub-benchmark runs in two orders: sorted iterates entities in creation
// order, random reshuffles them before every iteration (outside the timer).
func Benchmark_Editor_Add(b *testing.B) {
	ecs := setupECS()

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=1/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			var c1 goke.Comp[Pos]
			addEd := ecs.NewEditorBuilder(&c1).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos]()).Build()
			cur := &addEd.Cursor
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=2/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			var c1 goke.Comp[Pos]
			var c2 goke.Comp[Vel]
			addEd := ecs.NewEditorBuilder(&c1, &c2).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel]()).Build()
			cur := &addEd.Cursor
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=3/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			var c1 goke.Comp[Pos]
			var c2 goke.Comp[Vel]
			var c3 goke.Comp[Acc]
			addEd := ecs.NewEditorBuilder(&c1, &c2, &c3).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc]()).Build()
			cur := &addEd.Cursor
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=4/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			var c1 goke.Comp[Pos]
			var c2 goke.Comp[Vel]
			var c3 goke.Comp[Acc]
			var c4 goke.Comp[T04]
			addEd := ecs.NewEditorBuilder(&c1, &c2, &c3, &c4).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04]()).Build()
			cur := &addEd.Cursor
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=5/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
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
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=6/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
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
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=7/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
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
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=8/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
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
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=9/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
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
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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
	}

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/comps=10/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
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
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
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

	// Tag variants: same structural add, but the payload is N zero-size
	// tags — mask-only migration, no data columns to copy.
	ecs = setupTagECS()

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=1/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=2/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=3/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=4/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=5/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4]), new(goke.Comp[Tag5])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4](), goke.Del[Tag5]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=6/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4]), new(goke.Comp[Tag5]), new(goke.Comp[Tag6])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=7/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4]), new(goke.Comp[Tag5]), new(goke.Comp[Tag6]), new(goke.Comp[Tag7])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](), goke.Del[Tag7]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=8/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4]), new(goke.Comp[Tag5]), new(goke.Comp[Tag6]), new(goke.Comp[Tag7]), new(goke.Comp[Tag8])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](), goke.Del[Tag7](), goke.Del[Tag8]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=9/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4]), new(goke.Comp[Tag5]), new(goke.Comp[Tag6]), new(goke.Comp[Tag7]), new(goke.Comp[Tag8]), new(goke.Comp[Tag9])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](), goke.Del[Tag7](), goke.Del[Tag8](), goke.Del[Tag9]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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

	for _, o := range benchOrders {
		b.Run(fmt.Sprintf("pop=%d/subset=%d/comp=1/tags=10/%s", entitiesNumber, entitiesNumber, o), func(b *testing.B) {
			ecs.Reset()
			entities := populateBase(ecs, entitiesNumber)
			rng := rand.New(rand.NewPCG(42, 42))
			addEd := ecs.NewEditorBuilder(new(goke.Comp[Tag1]), new(goke.Comp[Tag2]), new(goke.Comp[Tag3]), new(goke.Comp[Tag4]), new(goke.Comp[Tag5]), new(goke.Comp[Tag6]), new(goke.Comp[Tag7]), new(goke.Comp[Tag8]), new(goke.Comp[Tag9]), new(goke.Comp[Tag10])).Build()
			delEd := ecs.NewEditorBuilder().Delete(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](), goke.Del[Tag7](), goke.Del[Tag8](), goke.Del[Tag9](), goke.Del[Tag10]()).Build()
			measurePerEntity(b, entitiesNumber, func() {
				for b.Loop() {
					if o == "random" {
						b.StopTimer()
						shuffleIDs(rng, entities)
						b.StartTimer()
					}
					for _, e := range entities {
						addEd.Update(e)
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
}
