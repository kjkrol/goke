package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
)

func Benchmark_View_All(b *testing.B) {

	ecs := setupECS()
	populate(ecs, entitiesNumber)
	var GlobalCount int
	b.Run("0 comp", func(b *testing.B) {
		view0 := goke.NewView0(ecs)
		fn := func() {
			count := 0
			for chunk := range view0.All() {
				for _, entityID := range chunk.Entity {
					_ = entityID
					count++
				}
			}
			GlobalCount = count
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})

		if GlobalCount != entitiesNumber {
			b.Fatalf("View0 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
		}
	})

	b.Run("1 comp", func(b *testing.B) {
		view1 := goke.NewView1[Pos](ecs)
		fn := func() {
			count := 0
			for chunk := range view1.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos := chunk.Comp1[i]
					pos.X += pos.Y
					count++
				}
			}
			GlobalCount = count
		}
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})

		if GlobalCount != entitiesNumber {
			b.Fatalf("View1 sanity check failed: expected %d, got %d", entitiesNumber, GlobalCount)
		}
	})

	b.Run("2 comp", func(b *testing.B) {
		view2 := goke.NewView2[Pos, Vel](ecs)
		fn := func() {
			for chunk := range view2.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel := chunk.Comp1[i], chunk.Comp2[i]
					vel.X += vel.Y
					pos.X += vel.X
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("3 comp", func(b *testing.B) {
		view3 := goke.NewView3[Pos, Vel, Acc](ecs, goke.Include[T04]())
		fn := func() {
			for chunk := range view3.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("4 comp", func(b *testing.B) {
		view4 := goke.NewView4[Pos, Vel, Acc, T04](ecs)
		fn := func() {
			for chunk := range view4.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("5 comp", func(b *testing.B) {
		view5 := goke.NewView5[Pos, Vel, Acc, T04, T05](ecs)
		fn := func() {
			for chunk := range view5.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
					chunk.Comp5[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("6 comp", func(b *testing.B) {
		view6 := goke.NewView6[Pos, Vel, Acc, T04, T05, T06](ecs)
		fn := func() {
			for chunk := range view6.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
					chunk.Comp5[i].V = 1
					chunk.Comp6[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("7 comp", func(b *testing.B) {
		view7 := goke.NewView7[Pos, Vel, Acc, T04, T05, T06, T07](ecs)
		fn := func() {
			for chunk := range view7.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
					chunk.Comp5[i].V = 1
					chunk.Comp6[i].V = 1
					chunk.Comp7[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("8 comp", func(b *testing.B) {
		view8 := goke.NewView8[Pos, Vel, Acc, T04, T05, T06, T07, T08](ecs)
		fn := func() {
			for chunk := range view8.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
					chunk.Comp5[i].V = 1
					chunk.Comp6[i].V = 1
					chunk.Comp7[i].V = 1
					chunk.Comp8[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("9 comp", func(b *testing.B) {
		view9 := goke.NewView9[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09](ecs)
		fn := func() {
			for chunk := range view9.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
					chunk.Comp5[i].V = 1
					chunk.Comp6[i].V = 1
					chunk.Comp7[i].V = 1
					chunk.Comp8[i].V = 1
					chunk.Comp9[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("10 comp", func(b *testing.B) {
		view10 := goke.NewView10[Pos, Vel, Acc, T04, T05, T06, T07, T08, T09, T10](ecs)
		fn := func() {
			for chunk := range view10.All() {
				for i, entityID := range chunk.Entity {
					_ = entityID
					pos, vel, acc := chunk.Comp1[i], chunk.Comp2[i], chunk.Comp3[i]
					acc.X += 0.1
					vel.X += acc.X
					pos.X += vel.X
					chunk.Comp4[i].V = 1
					chunk.Comp5[i].V = 1
					chunk.Comp6[i].V = 1
					chunk.Comp7[i].V = 1
					chunk.Comp8[i].V = 1
					chunk.Comp9[i].V = 1
					chunk.Comp10[i].V = 1
				}
			}
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})
}
