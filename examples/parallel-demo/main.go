package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kjkrol/goke/ecs"
)

// --- Components ---

type Dice struct{ Value int }
type Player struct{ Bet int }

// Winner is a tag component (empty struct) used to mark winning entities.
type Winner struct{}

// Global application state
var gameFinished = false
var turnCounter = 0

func main() {
	// 1. Initialize the engine
	engine := ecs.NewEngine()

	// Register component types (optimal for performance)
	winnerType := ecs.RegisterComponent[Winner](engine)

	// 2. Setup Entities & Components
	diceEnt := engine.CreateEntity()
	ecs.SetComponent(engine, diceEnt, Dice{Value: 0})

	// Setup player entities
	p1 := engine.CreateEntity()
	ecs.SetComponent(engine, p1, Player{Bet: 0})

	p2 := engine.CreateEntity()
	ecs.SetComponent(engine, p2, Player{Bet: 0})

	// 3. Define Views (for system filtering)
	vDice := ecs.NewView1[Dice](engine)
	vPlayers := ecs.NewView1[Player](engine)
	vWinners := ecs.NewView1[Winner](engine)

	// 4. Register Systems

	// System A: Roll the dice
	rollSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for res := range vDice.Values() {
			res.V1.Value = rand.Intn(6) + 1
		}
	})

	// System B: Players place their bets
	betSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for res := range vPlayers.Values() {
			res.V1.Bet = rand.Intn(6) + 1
		}
	})

	// System C: Judge the results
	judgeSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		if gameFinished {
			return
		}
		turnCounter++

		dice, _ := ecs.GetComponent[Dice](engine, diceEnt)
		fmt.Printf("🎲 Turn %d | Dice Result: %d\n", turnCounter, dice.Value)

		for res := range vPlayers.All() {
			fmt.Printf("   Player %d bet: %d\n", res.Entity, res.V1.Bet)
			if res.V1.Bet == dice.Value {
				gameFinished = true
				// Defer the assignment of the Winner tag to the next Sync point
				ecs.AssignComponent(cb, res.Entity, winnerType, Winner{})
			}
		}
	})

	// System D: Display winners (Reactive System)
	displayWinnerSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for res := range vWinners.All() {
			fmt.Printf("🏆 VICTORY! Entity %d is marked as a Winner!\n", res.Entity)
		}
	})

	// 5. Define Execution Plan
	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		// Run data updates in parallel
		ctx.RunParallel(d, rollSys, betSys)
		ctx.Sync()

		// Run judgment logic
		ctx.Run(judgeSys, d)

		// Crucial: Sync applies the Winner tag from judgeSys
		ctx.Sync()

		// Now the display system will see the entities in vWinners
		ctx.Run(displayWinnerSys, d)
	})

	// 6. Simulation Loop
	fmt.Println("Starting GOKe Dice Game Simulation...")
	for !gameFinished {
		engine.Tick(16 * time.Millisecond)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Simulation ended.")
}
