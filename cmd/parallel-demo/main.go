package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kjkrol/goke/pkg/ecs"
)

// --- Components ---

type Dice struct{ Value int }
type Player struct{ Bet int }
type GameState struct {
	Turn     int
	Finished bool
}

func main() {
	// Initialize the engine
	engine := ecs.NewEngine()

	// --- Entity & Component Setup ---

	// Setup dice entity
	diceEnt := engine.CreateEntity()
	ecs.SetComponent(engine, diceEnt, Dice{Value: 0})

	// Setup player entities
	p1 := engine.CreateEntity()
	ecs.SetComponent(engine, p1, Player{Bet: 0})

	p2 := engine.CreateEntity()
	ecs.SetComponent(engine, p2, Player{Bet: 0})

	// Setup global game state entity
	stateEnt := engine.CreateEntity()
	ecs.SetComponent(engine, stateEnt, GameState{Turn: 0, Finished: false})

	// --- Views ---
	// Views are defined here to be captured by system closures
	vDice := ecs.NewView1[Dice](engine)
	vPlayers := ecs.NewView1[Player](engine)

	// --- Systems Registration ---

	// System 1: Roll the dice
	rollSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for res := range vDice.Values() {
			res.V1.Value = rand.Intn(6) + 1
		}
	})

	// System 2: Players place their bets
	betSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		for res := range vPlayers.Values() {
			res.V1.Bet = rand.Intn(6) + 1
		}
	})

	// System 3: Judge the results and manage game state
	judgeSys := engine.RegisterSystemFunc(func(cb *ecs.Commands, d time.Duration) {
		// Access GameState via engine capture (closure)
		s, _ := ecs.GetComponent[GameState](engine, stateEnt)
		if s.Finished {
			return
		}
		s.Turn++

		// Get current dice value
		dice, _ := ecs.GetComponent[Dice](engine, diceEnt)
		fmt.Printf("🎲 Turn %d | Dice Result: %d\n", s.Turn, dice.Value)

		// Check all bets against the dice result
		for res := range vPlayers.All() {
			fmt.Printf("   Player %d bet: %d\n", res.Entity, res.V1.Bet)
			if res.V1.Bet == dice.Value {
				fmt.Printf("🏆 VICTORY! Player %d won the game!\n", res.Entity)
				s.Finished = true
			}
		}
	})

	// --- Execution Plan ---

	engine.SetExecutionPlan(func(ctx ecs.ExecutionContext, d time.Duration) {
		// Order matters: roll and bet first
		ctx.RunParallel(d, rollSys, betSys)
		ctx.Sync() // Apply changes to components

		// Then judge the state
		ctx.Run(judgeSys, d)
		ctx.Sync()
	})

	// --- Game Loop ---

	fmt.Println("Starting the dice game simulation...")
	for {
		// Check exit condition
		s, _ := ecs.GetComponent[GameState](engine, stateEnt)
		if s.Finished {
			break
		}

		// Advance engine state
		engine.Tick(16 * time.Millisecond)

		// Small delay for console readability
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("Simulation ended.")
}
