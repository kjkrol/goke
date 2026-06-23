package bench_test

import (
	"fmt"
	"testing"

	"github.com/kjkrol/goke"
)

type Tag1 struct{}
type Tag2 struct{}
type Tag3 struct{}
type Tag4 struct{}
type Tag5 struct{}
type Tag6 struct{}
type Tag7 struct{}
type Tag8 struct{}
type Tag9 struct{}
type Tag10 struct{}

func setupTagECS() *goke.ECS {
	ecs := setupECS()
	_ = goke.RegComp[Tag1](ecs)
	_ = goke.RegComp[Tag2](ecs)
	_ = goke.RegComp[Tag3](ecs)
	_ = goke.RegComp[Tag4](ecs)
	_ = goke.RegComp[Tag5](ecs)
	_ = goke.RegComp[Tag6](ecs)
	_ = goke.RegComp[Tag7](ecs)
	_ = goke.RegComp[Tag8](ecs)
	_ = goke.RegComp[Tag9](ecs)
	_ = goke.RegComp[Tag10](ecs)
	return ecs
}

// Benchmark_Editor_AddTags measures the per-entity cost of adding N zero-size
// tag components via Editor.Update. Entities carry only a Base anchor; each
// iteration adds N tags (timed) then resets back to Base outside the timer.
func Benchmark_Editor_AddTags(b *testing.B) {
	ecs := setupTagECS()

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=1", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])))
		delEd := ecs.CreateEditor(goke.Del[Tag1]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=2", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=3", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=4", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](), goke.Del[Tag4]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=5", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])),
			goke.Add(new(goke.Col[Tag5])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
			goke.Del[Tag4](), goke.Del[Tag5]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=6", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])),
			goke.Add(new(goke.Col[Tag5])), goke.Add(new(goke.Col[Tag6])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
			goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=7", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])),
			goke.Add(new(goke.Col[Tag5])), goke.Add(new(goke.Col[Tag6])),
			goke.Add(new(goke.Col[Tag7])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
			goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](), goke.Del[Tag7]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=8", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])),
			goke.Add(new(goke.Col[Tag5])), goke.Add(new(goke.Col[Tag6])),
			goke.Add(new(goke.Col[Tag7])), goke.Add(new(goke.Col[Tag8])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
			goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](),
			goke.Del[Tag7](), goke.Del[Tag8]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=9", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])),
			goke.Add(new(goke.Col[Tag5])), goke.Add(new(goke.Col[Tag6])),
			goke.Add(new(goke.Col[Tag7])), goke.Add(new(goke.Col[Tag8])),
			goke.Add(new(goke.Col[Tag9])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
			goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](),
			goke.Del[Tag7](), goke.Del[Tag8](), goke.Del[Tag9]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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

	b.Run(fmt.Sprintf("pop=%d/comp=1/tags=10", entitiesNumber), func(b *testing.B) {
		ecs.Reset()
		entities := populateBase(ecs, entitiesNumber)
		addEd := ecs.CreateEditor(goke.Add(new(goke.Col[Tag1])), goke.Add(new(goke.Col[Tag2])),
			goke.Add(new(goke.Col[Tag3])), goke.Add(new(goke.Col[Tag4])),
			goke.Add(new(goke.Col[Tag5])), goke.Add(new(goke.Col[Tag6])),
			goke.Add(new(goke.Col[Tag7])), goke.Add(new(goke.Col[Tag8])),
			goke.Add(new(goke.Col[Tag9])), goke.Add(new(goke.Col[Tag10])))
		delEd := ecs.CreateEditor(goke.Del[Tag1](), goke.Del[Tag2](), goke.Del[Tag3](),
			goke.Del[Tag4](), goke.Del[Tag5](), goke.Del[Tag6](),
			goke.Del[Tag7](), goke.Del[Tag8](), goke.Del[Tag9](), goke.Del[Tag10]())
		measurePerEntity(b, entitiesNumber, func() {
			for b.Loop() {
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
