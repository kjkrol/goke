package ecs

import (
	"time"
)

type scheduler struct {
	register *Registry
	systems  []System
}

func newScheduler(register *Registry) *scheduler {
	return &scheduler{
		register: register,
		systems:  make([]System, 0),
	}
}

func (e *scheduler) registerSystems(systems []System) {
	for _, system := range systems {
		system.Init(e.register)
		e.systems = append(e.systems, system)
	}
}

func (e *scheduler) updateSystems(duration time.Duration) {
	for _, system := range e.systems {
		system.Update(e.register, duration)
	}
}
