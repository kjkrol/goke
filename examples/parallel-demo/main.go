package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kjkrol/goke"
)

// --- Components ---

type Dice struct{ Value int }
type Player struct{ Bet int }

// Winner is a tag component (empty struct) used to mark winning entities.
type Winner struct{}

// Global application state
var gameFinished = false
var turnCounter = 0

var winnerDesc, diceDesc, playerDesc goke.ComponentDesc

func main() {
	// 1. Initialize the ecs
	ecs := goke.New()

	// Register component types
	winnerDesc = goke.RegisterComponent[Winner](ecs)
	diceDesc = goke.RegisterComponent[Dice](ecs)
	playerDesc = goke.RegisterComponent[Player](ecs)

	// 2. Setup Entities & Components
	diceBlueprint := goke.NewBlueprint1[Dice](ecs)

	var diceEnt goke.Entity
	for page := range diceBlueprint.Create(1) {
		diceEnt = page.Entity[0]
		page.Comp1[0] = Dice{Value: 0}
	}

	// Setup player entities
	playerBlueprint := goke.NewBlueprint1[Player](ecs)

	for page := range playerBlueprint.Create(2) {
		for i, _ := range page.Entity {
			page.Comp1[i] = Player{Bet: 0}
		}
	}

	// 3. Define Views (for system filtering)
	vDice := goke.NewView1[Dice](ecs)
	vPlayers := goke.NewView1[Player](ecs)
	vWinners := goke.NewView0(ecs, goke.Include[Winner]())

	// 4. Register Systems

	// System A: Roll the dice
	rollSys := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for page := range vDice.All() {
			for i, _ := range page.Entity {
				page.Comp1[i].Value = rand.Intn(6) + 1
			}
		}
	})

	// System B: Players place their bets
	betSys := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for page := range vPlayers.All() {
			for i, _ := range page.Entity {
				page.Comp1[i].Bet = rand.Intn(6) + 1
			}
		}
	})

	// System C: Judge the results
	judgeSys := goke.RegisterSystemFunc(ecs, func(schedule *goke.Schedule, d time.Duration) {
		if gameFinished {
			return
		}
		turnCounter++

		dice, _ := goke.SafeGetComponent[Dice](ecs, diceEnt, diceDesc)
		fmt.Printf("🎲 Turn %d | Dice Result: %d\n", turnCounter, dice.Value)

		for page := range vPlayers.All() {
			for i, entity := range page.Entity {
				bet := page.Comp1[i].Bet
				fmt.Printf("   Player %d bet: %d\n", entity, bet)
				if bet == dice.Value {
					gameFinished = true
					// Defer the assignment of the Winner tag to the next Sync point
					goke.ScheduleAddComponent(schedule, entity, winnerDesc, Winner{})
				}
			}
		}
	})

	// System D: Display winners (Reactive System)
	displayWinnerSys := goke.RegisterSystemFunc(ecs, func(cb *goke.Schedule, d time.Duration) {
		for res := range vWinners.All() {
			fmt.Printf("🏆 VICTORY! Entity %d is marked as a Winner!\n", res.Entity)
		}
	})

	// 5. Define Execution Plan
	goke.Plan(ecs, func(ctx goke.ExecutionContext, d time.Duration) {
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
		goke.Tick(ecs, 16*time.Millisecond)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Simulation ended.")
}
