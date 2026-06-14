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

var winnerDesc, diceDesc, playerDesc goke.CompMeta

func main() {
	// 1. Initialize the ecs
	ecs := goke.New()

	// Register component types
	winnerDesc = goke.RegCompType[Winner](ecs)
	diceDesc = goke.RegCompType[Dice](ecs)
	playerDesc = goke.RegCompType[Player](ecs)

	// 2. Setup Entities & Components
	diceBlueprint := goke.NewBlueprint1[Dice](ecs)

	var diceEnt goke.EntityID
	for chunk := range diceBlueprint.Create(1) {
		diceEnt = chunk.Entity[0]
		chunk.Comp1[0] = Dice{Value: 0}
	}

	// Setup player entities
	playerBlueprint := goke.NewBlueprint1[Player](ecs)

	for chunk := range playerBlueprint.Create(2) {
		for i, _ := range chunk.Entity {
			chunk.Comp1[i] = Player{Bet: 0}
		}
	}

	// 3. Define Views (for system filtering)
	vDice := goke.NewView1[Dice](ecs)
	vPlayers := goke.NewView1[Player](ecs)
	vWinners := goke.NewView0(ecs, goke.Include[Winner]())

	// 4. Register Systems

	// System A: Roll the dice
	rollSys := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		for chunk := range vDice.All() {
			for i, _ := range chunk.Entity {
				chunk.Comp1[i].Value = rand.Intn(6) + 1
			}
		}
	})

	// System B: Players place their bets
	betSys := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		for chunk := range vPlayers.All() {
			for i, _ := range chunk.Entity {
				chunk.Comp1[i].Bet = rand.Intn(6) + 1
			}
		}
	})

	// System C: Judge the results
	judgeSys := goke.RegSysFn(ecs, func(schedule *goke.CmdBuf, d time.Duration) {
		if gameFinished {
			return
		}
		turnCounter++

		dice, _ := goke.SafeGetComp[Dice](ecs, diceEnt, diceDesc)
		fmt.Printf("🎲 Turn %d | Dice Result: %d\n", turnCounter, dice.Value)

		for chunk := range vPlayers.All() {
			for i, entityID := range chunk.Entity {
				bet := chunk.Comp1[i].Bet
				fmt.Printf("   Player %d bet: %d\n", entityID, bet)
				if bet == dice.Value {
					gameFinished = true
					// Defer the assignment of the Winner tag to the next Sync point
					goke.CmdBufAddComp(schedule, entityID, winnerDesc, Winner{})
				}
			}
		}
	})

	// System D: Display winners (Reactive System)
	displayWinnerSys := goke.RegSysFn(ecs, func(cb *goke.CmdBuf, d time.Duration) {
		for res := range vWinners.All() {
			fmt.Printf("🏆 VICTORY! Entity %d is marked as a Winner!\n", res.Entity)
		}
	})

	// 5. Define Execution Plan
	goke.SetPlan(ecs, func(ctx goke.RunCtx, d time.Duration) {
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
