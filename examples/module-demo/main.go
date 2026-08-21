package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Position and Velocity are movementModule's own components — a caller
// wiring up the module never names them directly.
type Position struct{ X, Y float32 }
type Velocity struct{ X, Y float32 }

// movementModule is a self-contained package: it owns Position/Velocity,
// seeds a batch of entities at Setup, and advances them every tick.
type movementModule struct {
	count int
	pos   goke.Comp[Position]
	vel   goke.Comp[Velocity]
	query *goke.Query
	sys   goke.Runnable
}

func newMovementModule(count int) *movementModule { return &movementModule{count: count} }

func (m *movementModule) SetupSystems() []goke.System {
	return []goke.System{goke.SystemFn{OnInit: func(si *goke.SysInit) {
		si.RegComp[Position]()
		si.RegComp[Velocity]()

		factory := si.NewFactory(&m.pos, &m.vel)
		factory.Create(m.count)
		for factory.Next() {
			positions := m.pos.Slice(&factory.Cursor)
			velocities := m.vel.Slice(&factory.Cursor)
			for i := range positions {
				positions[i] = Position{}
				velocities[i] = Velocity{X: rand.Float32()*2 - 1, Y: rand.Float32()*2 - 1}
			}
		}
	}}}
}

func (m *movementModule) RegSystems(ecs *goke.ECS) {
	m.sys = ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			m.query = si.NewQueryBuilder(&m.pos, &m.vel).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			cursor := m.query.Cursor()
			m.query.All()
			for m.query.Next() {
				positions := m.pos.Slice(cursor)
				velocities := m.vel.Slice(cursor)
				for i := range positions {
					positions[i].X += velocities[i].X
					positions[i].Y += velocities[i].Y
				}
			}
		},
	})
}

func (m *movementModule) RunPlan(ctx goke.RunCtx, d time.Duration) { ctx.Run(m.sys, d) }

func (m *movementModule) LoadComps() []goke.CompToken {
	return goke.LoadComps(&m.pos, &m.vel)
}

var _ goke.Module = (*movementModule)(nil)

// Tally is statsModule's own component. statsModule tracks it on a single
// entity it seeds for itself.
type Tally struct{ Ticks, Moving int }

// statsModule is a second, independent module: it doesn't know
// movementModule exists as a Go value, only that something may have
// registered the public Position component type — decoupled by component
// type, not by a direct reference to movementModule.
type statsModule struct {
	tally       goke.Comp[Tally]
	tallyID     uid.UID64
	query       *goke.Query
	movingQuery *goke.Query
	sys         goke.Runnable
}

func newStatsModule() *statsModule { return &statsModule{} }

func (s *statsModule) SetupSystems() []goke.System {
	return []goke.System{goke.SystemFn{OnInit: func(si *goke.SysInit) {
		si.RegComp[Tally]()

		factory := si.NewFactory(&s.tally)
		factory.Create(1)
		factory.Next()
		s.tallyID = factory.IDs[0]
		s.tally.Slice(&factory.Cursor)[0] = Tally{}
	}}}
}

func (s *statsModule) RegSystems(ecs *goke.ECS) {
	s.sys = ecs.RegSys(goke.SystemFn{
		OnInit: func(si *goke.SysInit) {
			s.query = si.NewQueryBuilder(&s.tally).Build()
			s.movingQuery = si.NewQueryBuilder().Include(goke.Include[Position]()).Build()
		},
		OnUpdate: func(_ *goke.CmdBuf, _ time.Duration) {
			moving := 0
			s.movingQuery.All()
			for s.movingQuery.Next() {
				moving += len(s.movingQuery.Cursor().IDs)
			}
			if s.query.Seek(s.tallyID) {
				t := s.tally.At(s.query.Cursor())
				t.Ticks++
				t.Moving = moving
			}
		},
	})
}

func (s *statsModule) RunPlan(ctx goke.RunCtx, d time.Duration) { ctx.Run(s.sys, d) }

func (s *statsModule) LoadComps() []goke.CompToken {
	return goke.LoadComps(&s.tally)
}

// Report exposes statsModule's tally without leaking its Query/Comp
// internals to the caller.
func (s *statsModule) Report() Tally {
	if s.query.Seek(s.tallyID) {
		return *s.tally.At(s.query.Cursor())
	}
	return Tally{}
}

var _ goke.Module = (*statsModule)(nil)

func main() {
	ecs := goke.New()

	movement := newMovementModule(5)
	stats := newStatsModule()

	ecs.RegModule(movement)
	ecs.RegModule(stats)

	ecs.Setup(append(movement.SetupSystems(), stats.SetupSystems()...)...)

	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		movement.RunPlan(ctx, d)
		ctx.Sync()
		stats.RunPlan(ctx, d)
	})

	for range 5 {
		ecs.Tick(time.Second)
	}

	tally := stats.Report()
	fmt.Printf("ticks=%d entities-with-Position=%d\n", tally.Ticks, tally.Moving)
}
