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
	query          *ecs.CachedQuery3[Order, Status, Discount]
	processedCount int
}

func (s *BillingSystem) Init(reg *ecs.Registry) {
	s.query = ecs.NewQuery3[Order, Status, Discount](reg)
	s.processedCount = 0
}

func (s *BillingSystem) Update(reg *ecs.Registry, d time.Duration) {
	for _, row := range s.query.All() {
		s.processedCount++
		ord := row.V1
		st := row.V2
		disc := row.V3
		ord.Total = ord.Total * (1 - disc.Percentage/100)
		st.Processed = true
	}
}

var _ ecs.System = (*BillingSystem)(nil)

func TestECS_UseCase(t *testing.T) {
	engine := ecs.NewEngine()

	// Encja A: Spełnia wymagania (Order + Status + Discount)
	eA := engine.CreateEntity()
	ecs.Assign(engine, eA, Order{ID: "ORD-001", Total: 100.0})
	ecs.Assign(engine, eA, Status{Processed: false})
	ecs.Assign(engine, eA, Discount{Percentage: 10.0})

	// Encja B: Nie spełnia wymagań (brak Discount)
	eB := engine.CreateEntity()
	ecs.Assign(engine, eB, Order{ID: "ORD-002", Total: 50.0})
	ecs.Assign(engine, eB, Status{Processed: false})

	// 3. System Przetwarzania
	billingSystem := BillingSystem{}
	engine.RegisterSystems([]ecs.System{&billingSystem})

	engine.UpdateSystems(time.Duration(time.Second))

	// Sprawdź, czy Each znalazł tylko 1 encję (eA)
	if billingSystem.processedCount != 1 {
		t.Errorf("System powinien przetworzyć 1 encję, przetworzył %d", billingSystem.processedCount)
	}

	// Sprawdź, czy wskaźnik zadziałał (czy dane w Registry się zmieniły)
	order := ecs.Get[Order](engine, eA)
	status := ecs.Get[Status](engine, eA)

	if order.Total != 90.0 {
		t.Errorf("Rabat nie został naliczony poprawnie, Total: %v", order.Total)
	}

	if !status.Processed {
		t.Error("Status nie został zmieniony na Processed")
	}

	engine.RemoveEntity(eA)

	// Sprawdź czy dane fizycznie zniknęły z mapy storage (sprzątanie deleterem)
	if ecs.Get[Order](engine, eA) != nil {
		t.Error("Dane encji A powinny zostać usunięte z mapy Order")
	}
}
