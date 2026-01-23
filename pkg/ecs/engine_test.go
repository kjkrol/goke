package ecs_test

import (
	"testing"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
)

// components

type Order struct {
	ID    string
	Total float64
}

type Status struct {
	Processed bool
}

type Discount struct {
	Percentage float64
}

// system

type BillingSystem struct {
	query          *ecs.Query3[Order, Status, Discount]
	processedCount int
}

func (s *BillingSystem) Init(reg *ecs.Registry) {
	s.query = ecs.NewQuery3[Order, Status, Discount](reg)
	s.processedCount = 0
}

func (s *BillingSystem) Update(reg *ecs.Registry, d time.Duration) {
	for head := range s.query.All3() {
		s.processedCount++
		ord, st, disc := head.V1, head.V2, head.V3
		ord.Total = ord.Total * (1 - disc.Percentage/100)
		st.Processed = true
	}
}

var _ ecs.System = (*BillingSystem)(nil)

func TestECS_UseCase(t *testing.T) {
	engine := ecs.NewEngine()

	eA := engine.CreateEntity()
	ecs.Assign(engine, eA, Order{ID: "ORD-001", Total: 100.0})
	ecs.Assign(engine, eA, Status{Processed: false})
	ecs.Assign(engine, eA, Discount{Percentage: 10.0})

	eB := engine.CreateEntity()
	ecs.Assign(engine, eB, Order{ID: "ORD-002", Total: 50.0})
	ecs.Assign(engine, eB, Status{Processed: false})

	billingSystem := BillingSystem{}
	engine.RegisterSystems([]ecs.System{&billingSystem})

	engine.UpdateSystems(time.Duration(time.Second))

	if billingSystem.processedCount != 1 {
		t.Errorf("BillingSystem should have processed 1 entity, but it processed %d",
			billingSystem.processedCount)
	}

	order, _ := ecs.GetComponent[Order](engine, eA)
	status, _ := ecs.GetComponent[Status](engine, eA)

	if order.Total != 90.0 {
		t.Errorf("Discount has not been applied, Total: %v", order.Total)
	}

	if !status.Processed {
		t.Error("Status has not been changed to Processed")
	}

	engine.RemoveEntity(eA)

	_, err := ecs.GetComponent[Order](engine, eA)

	if err == nil {
		t.Error("Data of entity eA should have been removed from orders map")
	}
}
