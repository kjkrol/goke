package bench_test

import (
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/goke/internal/core"
)

func Benchmark_View_Each(b *testing.B) {

	ecs := setupECS()
	populate(ecs, entitiesNumber)

	b.Run("1 comp", func(b *testing.B) {
		view1 := goke.NewView1[Pos](ecs)
		fn := func() {
			view1.Each(func(
				entities []core.Entity,
				pos []Pos,
			) {
				for i := range entities {
					pos[i].X += pos[i].Y
				}
			})
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("2 comp", func(b *testing.B) {
		view2 := goke.NewView2[Pos, Vel](ecs)
		fn := func() {
			view2.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
			) {
				for i := range entities {
					vel[i].X += vel[i].Y
					pos[i].X += vel[i].X
				}
			})
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})

	b.Run("3 comp", func(b *testing.B) {
		view3 := goke.NewView3[Pos, Vel, Acc](ecs)
		fn := func() {
			view3.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
				}
			})
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
			view4.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
				}
			})
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
			view5.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
				t05 []T05,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
					t05[i].V = 1
				}
			})
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
			view6.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
				t05 []T05,
				t06 []T06,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
					t05[i].V = 1
					t06[i].V = 1
				}
			})
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
			view7.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
				t05 []T05,
				t06 []T06,
				t07 []T07,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
					t05[i].V = 1
					t06[i].V = 1
					t07[i].V = 1
				}
			})
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
			view8.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
				t05 []T05,
				t06 []T06,
				t07 []T07,
				t08 []T08,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
					t05[i].V = 1
					t06[i].V = 1
					t07[i].V = 1
					t08[i].V = 1
				}
			})
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
			view9.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
				t05 []T05,
				t06 []T06,
				t07 []T07,
				t08 []T08,
				t09 []T09,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
					t05[i].V = 1
					t06[i].V = 1
					t07[i].V = 1
					t08[i].V = 1
					t09[i].V = 1
				}
			})
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
			view10.Each(func(
				entities []core.Entity,
				pos []Pos,
				vel []Vel,
				acc []Acc,
				t04 []T04,
				t05 []T05,
				t06 []T06,
				t07 []T07,
				t08 []T08,
				t09 []T09,
				t10 []T10,
			) {
				for i := range entities {
					acc[i].X += 0.1
					vel[i].X += acc[i].X
					pos[i].X += vel[i].X
					t04[i].V = 1
					t05[i].V = 1
					t06[i].V = 1
					t07[i].V = 1
					t08[i].V = 1
					t09[i].V = 1
					t10[i].V = 1
				}
			})
		}

		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				fn()
			}
		})
	})
}
