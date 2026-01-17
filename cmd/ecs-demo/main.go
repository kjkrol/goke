package main

import (
	"fmt"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
)

type (
	Order struct {
		ID    string
		Total float64
	}
	Status   struct{ Processed bool }
	Discount struct{ Percentage float64 }

	BillingSystem struct {
		view *ecs.View3[Order, Status, Discount]
	}
)

func (s *BillingSystem) Init(reg *ecs.Registry) {
	s.view = ecs.NewView3[Order, Status, Discount](reg)
}

func (s *BillingSystem) Update(reg *ecs.Registry, d time.Duration) {
	for head, tail := range ecs.All3(s.view) {
		ord, st, disc := head.V1, head.V2, tail.V3
		ord.Total = ord.Total * (1 - disc.Percentage/100)
		st.Processed = true
	}
}

var _ ecs.System = (*BillingSystem)(nil)

func main() {
	engine := ecs.NewEngine()

	entity := engine.CreateEntity()
	ecs.Assign(engine, entity, Order{ID: "ORD-99", Total: 200.0})
	ecs.Assign(engine, entity, Status{Processed: false})
	ecs.Assign(engine, entity, Discount{Percentage: 20.0})

	billing := &BillingSystem{}
	engine.RegisterSystems([]ecs.System{billing})

	order, _ := ecs.GetComponent[Order](engine, entity)
	fmt.Printf("Order id: %v value: %v\n", order.ID, order.Total)
	engine.UpdateSystems(time.Duration(time.Second))
	fmt.Printf("Order id: %v value with discount: %v\n", order.ID, order.Total)
}
