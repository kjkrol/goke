package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
	"github.com/kjkrol/uid"
)

// E01..E10 are the incoming component types used by Benchmark_Editor_Mix.
// For N=K the editor adds E01..EK and removes the K existing components,
// replacing them completely in a single archetype migration.
type E01 struct{ V float32 }
type E02 struct{ V float32 }
type E03 struct{ V float32 }
type E04 struct{ V float32 }
type E05 struct{ V float32 }
type E06 struct{ V float32 }
type E07 struct{ V float32 }
type E08 struct{ V float32 }
type E09 struct{ V float32 }
type E10 struct{ V float32 }

func setupEditECS() *goke.ECS {
	ecs := setupECS()
	_ = goke.RegComp[E01](ecs)
	_ = goke.RegComp[E02](ecs)
	_ = goke.RegComp[E03](ecs)
	_ = goke.RegComp[E04](ecs)
	_ = goke.RegComp[E05](ecs)
	_ = goke.RegComp[E06](ecs)
	_ = goke.RegComp[E07](ecs)
	_ = goke.RegComp[E08](ecs)
	_ = goke.RegComp[E09](ecs)
	_ = goke.RegComp[E10](ecs)
	return ecs
}

func populateN(ecs *goke.ECS, count, n int) []uid.UID64 {
	var ids []uid.UID64
	switch n {
	case 1:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 2:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 3:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 4:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 5:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 6:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])),
			goke.Add(new(goke.Col[T06])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 7:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])),
			goke.Add(new(goke.Col[T06])), goke.Add(new(goke.Col[T07])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 8:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])),
			goke.Add(new(goke.Col[T06])), goke.Add(new(goke.Col[T07])),
			goke.Add(new(goke.Col[T08])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 9:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])),
			goke.Add(new(goke.Col[T06])), goke.Add(new(goke.Col[T07])),
			goke.Add(new(goke.Col[T08])), goke.Add(new(goke.Col[T09])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	case 10:
		f := ecs.CreateFactory(goke.Add(new(goke.Col[Base])), goke.Add(new(goke.Col[Pos])),
			goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])),
			goke.Add(new(goke.Col[T06])), goke.Add(new(goke.Col[T07])),
			goke.Add(new(goke.Col[T08])), goke.Add(new(goke.Col[T09])),
			goke.Add(new(goke.Col[T10])))
		f.Create(count)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
	}
	return ids
}

// Benchmark_Editor_Mix measures the per-entity cost of a combined add+remove
// via Editor.Update. Each sub-benchmark N adds N new types (E01..EN) and
// removes N existing ones (Pos..T0N) in a single migration (timed). A reverse
// editor undoes the migration outside the timer so every loop iteration starts
// from the same archetype.
func Benchmark_Editor_Mix(b *testing.B) {
	ecs := setupEditECS()

	b.Run(fmt.Sprintf("pop=%d/comp=2/swap=1", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 1)
		var a1 goke.Col[E01]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Del[Pos]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Del[E01]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=3/swap=2", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 2)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2),
			goke.Del[Pos](), goke.Del[Vel]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])),
			goke.Del[E01](), goke.Del[E02]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=4/swap=3", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 3)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=5/swap=4", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 4)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3), goke.Add(&a4),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])),
			goke.Add(new(goke.Col[Acc])), goke.Add(new(goke.Col[T04])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](), goke.Del[E04]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=6/swap=5", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 5)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		var a5 goke.Col[E05]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3), goke.Add(&a4), goke.Add(&a5),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](), goke.Del[T05]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](), goke.Del[E04](), goke.Del[E05]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
					a5.At(cur).V = 5
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=7/swap=6", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 6)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		var a5 goke.Col[E05]
		var a6 goke.Col[E06]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3),
			goke.Add(&a4), goke.Add(&a5), goke.Add(&a6),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](),
			goke.Del[E04](), goke.Del[E05](), goke.Del[E06]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
					a5.At(cur).V = 5
					a6.At(cur).V = 6
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=8/swap=7", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 7)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		var a5 goke.Col[E05]
		var a6 goke.Col[E06]
		var a7 goke.Col[E07]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3),
			goke.Add(&a4), goke.Add(&a5), goke.Add(&a6), goke.Add(&a7),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](), goke.Del[T07]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](),
			goke.Del[E04](), goke.Del[E05](), goke.Del[E06](), goke.Del[E07]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
					a5.At(cur).V = 5
					a6.At(cur).V = 6
					a7.At(cur).V = 7
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=9/swap=8", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 8)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		var a5 goke.Col[E05]
		var a6 goke.Col[E06]
		var a7 goke.Col[E07]
		var a8 goke.Col[E08]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3), goke.Add(&a4),
			goke.Add(&a5), goke.Add(&a6), goke.Add(&a7), goke.Add(&a8),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](), goke.Del[T04](),
			goke.Del[T05](), goke.Del[T06](), goke.Del[T07](), goke.Del[T08]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])), goke.Add(new(goke.Col[T08])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](), goke.Del[E04](),
			goke.Del[E05](), goke.Del[E06](), goke.Del[E07](), goke.Del[E08]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
					a5.At(cur).V = 5
					a6.At(cur).V = 6
					a7.At(cur).V = 7
					a8.At(cur).V = 8
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=10/swap=9", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 9)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		var a5 goke.Col[E05]
		var a6 goke.Col[E06]
		var a7 goke.Col[E07]
		var a8 goke.Col[E08]
		var a9 goke.Col[E09]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3),
			goke.Add(&a4), goke.Add(&a5), goke.Add(&a6),
			goke.Add(&a7), goke.Add(&a8), goke.Add(&a9),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08](), goke.Del[T09]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])), goke.Add(new(goke.Col[T08])), goke.Add(new(goke.Col[T09])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](),
			goke.Del[E04](), goke.Del[E05](), goke.Del[E06](),
			goke.Del[E07](), goke.Del[E08](), goke.Del[E09]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
					a5.At(cur).V = 5
					a6.At(cur).V = 6
					a7.At(cur).V = 7
					a8.At(cur).V = 8
					a9.At(cur).V = 9
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})

	b.Run(fmt.Sprintf("pop=%d/comp=11/swap=10", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateN(ecs, entitiesNumber, 10)
		var a1 goke.Col[E01]
		var a2 goke.Col[E02]
		var a3 goke.Col[E03]
		var a4 goke.Col[E04]
		var a5 goke.Col[E05]
		var a6 goke.Col[E06]
		var a7 goke.Col[E07]
		var a8 goke.Col[E08]
		var a9 goke.Col[E09]
		var a10 goke.Col[E10]
		fwdEd := ecs.CreateEditor(goke.Add(&a1), goke.Add(&a2), goke.Add(&a3),
			goke.Add(&a4), goke.Add(&a5), goke.Add(&a6),
			goke.Add(&a7), goke.Add(&a8), goke.Add(&a9), goke.Add(&a10),
			goke.Del[Pos](), goke.Del[Vel](), goke.Del[Acc](),
			goke.Del[T04](), goke.Del[T05](), goke.Del[T06](),
			goke.Del[T07](), goke.Del[T08](), goke.Del[T09](), goke.Del[T10]())
		revEd := ecs.CreateEditor(goke.Add(new(goke.Col[Pos])), goke.Add(new(goke.Col[Vel])), goke.Add(new(goke.Col[Acc])),
			goke.Add(new(goke.Col[T04])), goke.Add(new(goke.Col[T05])), goke.Add(new(goke.Col[T06])),
			goke.Add(new(goke.Col[T07])), goke.Add(new(goke.Col[T08])), goke.Add(new(goke.Col[T09])),
			goke.Add(new(goke.Col[T10])),
			goke.Del[E01](), goke.Del[E02](), goke.Del[E03](),
			goke.Del[E04](), goke.Del[E05](), goke.Del[E06](),
			goke.Del[E07](), goke.Del[E08](), goke.Del[E09](), goke.Del[E10]())
		cur := &fwdEd.Cursor
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
				for _, e := range entities {
					fwdEd.Update(e)
					a1.At(cur).V = 1
					a2.At(cur).V = 2
					a3.At(cur).V = 3
					a4.At(cur).V = 4
					a5.At(cur).V = 5
					a6.At(cur).V = 6
					a7.At(cur).V = 7
					a8.At(cur).V = 8
					a9.At(cur).V = 9
					a10.At(cur).V = 10
				}
				b.StopTimer()
				for _, e := range entities {
					revEd.Update(e)
				}
				b.StartTimer()
			}
		})
	})
}
