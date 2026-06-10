package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/kjkrol/uid"
)

type mockCompA struct {
	Val int
}

type mockCompB struct {
	Msg string
}

type modifyTestSystem struct {
	compA  ComponentInfo
	compB  ComponentInfo
	target uid.UID64
}

func (s *modifyTestSystem) Update(reg ReadOnlyRegistry, cb *SystemCommandBuffer, d time.Duration) {
	// Mutujemy już wygenerowaną i przydzieloną encję
	RemoveComponent(cb, s.target, s.compA)
	AddComponent(cb, s.target, s.compB, mockCompB{Msg: "added"})
}

func TestScheduler_ComponentCommands(t *testing.T) {
	reg := NewRegistry(RegistryConfig{
		InitialEntityCap:            100,
		InitialArchetypeRegistryCap: 100,
		FreeIndicesCap:              100,
		ViewRegistryInitCap:         10,
	})

	compA := reg.RegisterComponentType(reflect.TypeFor[mockCompA]())
	compB := reg.RegisterComponentType(reflect.TypeFor[mockCompB]())

	sched := NewScheduler(reg)

	// Symulacja prealokacji: Object Pooling
	e := reg.CreateEntity()
	ptrA, _ := reg.AllocateByID(e, compA)
	*(*mockCompA)(ptrA) = mockCompA{Val: 100}
	reg.AllocateByID(e, compB) // Rejestrujemy oba by UnAssign zadziałało

	sys := &modifyTestSystem{
		compA:  compA,
		compB:  compB,
		target: e,
	}

	sched.RegisterSystem(sys)
	sched.Run(sys, 0)

	err := sched.Sync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Powinno usunąć Component A
	_, err = reg.ComponentGet(e, compA.ID)
	if err == nil {
		t.Errorf("Expected compA to be removed")
	}

	// Powinno nadpisać Component B
	ptrB, err := reg.ComponentGet(e, compB.ID)
	if err != nil {
		t.Fatalf("Expected compB to be present")
	}
	if (*mockCompB)(ptrB).Msg != "added" {
		t.Errorf("Expected compB.Msg to be 'added'")
	}
}

func TestSystemCommandBuffer_Clear(t *testing.T) {
	cb := NewSystemCommandBuffer()

	RemoveEntity(cb, uid.UID64(1))

	if len(cb.commands) == 0 {
		t.Fatalf("Expected commands to not be empty")
	}

	cb.Clear()

	if len(cb.commands) != 0 {
		t.Errorf("Expected commands to be empty")
	}
}

func TestSystemCommandBuffer_ReserveSpace(t *testing.T) {
	cb := NewSystemCommandBuffer()

	p1 := cb.reserveSpace(10, 1)
	if p1 == nil {
		t.Errorf("Expected non-nil pointer")
	}

	p2 := cb.reserveSpace(8192, 1)
	if p2 == nil {
		t.Errorf("Expected non-nil pointer for large alloc")
	}
}
