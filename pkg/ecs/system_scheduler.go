package ecs

import (
	"time"
)

type SystemScheduler struct {
	register *Registry
	systems  []System
	cmdBuf   *SystemCommandBuffer
}

func newScheduler(register *Registry) *SystemScheduler {
	return &SystemScheduler{
		register: register,
		systems:  make([]System, 0),
		cmdBuf:   NewSystemCommandBuffer(),
	}
}

func (e *SystemScheduler) registerSystems(systems []System) {
	for _, system := range systems {
		system.Init(e.register)
		e.systems = append(e.systems, system)
	}
}

func (e *SystemScheduler) updateSystems(duration time.Duration) {
	for _, system := range e.systems {
		system.Update(e.register, e.cmdBuf, duration)

		if system.ShouldSync() {
			e.applyCommands()
		}
	}
	e.applyCommands()

}

func (e *SystemScheduler) applyCommands() {
	for _, cmd := range e.cmdBuf.commands {
		switch cmd.cType {
		case cmdAssignComponent:
			e.register.AssignByID(cmd.entity, cmd.compID, cmd.dataPtr)

		case cmdRemoveComponent:
			e.register.UnassignByID(cmd.entity, cmd.compID)

		case cmdRemoveEntity:
			e.register.RemoveEntity(cmd.entity)
		}
	}
	e.cmdBuf.reset()
}
