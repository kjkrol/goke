package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kjkrol/goke/v2"
	"github.com/kjkrol/uid"
)

// --- Components ---

type Dice struct{ Value int }
type Player struct{ Bet int }

// Winner is a tag component (empty struct) used to mark winning entities.
type Winner struct{}

// Global application state
var gameFinished = false
var turnCounter = 0

var winnerID, diceID, playerID goke.CompID

func main() {
	// 1. Initialize the ecs
	ecs := goke.New()

	// Register component types
	winnerID = goke.RegComp[Winner](ecs)
	diceID = goke.RegComp[Dice](ecs)
	playerID = goke.RegComp[Player](ecs)

	// 2. Setup Entities & Components
	var dice goke.Comp[Dice]
	diceFactory := ecs.NewFactory(&dice)

	var diceEnt uid.UID64
	diceFactory.Create(1)
	diceFactory.Next()
	diceEnt = diceFactory.IDs[0]
	dice.Slice(&diceFactory.Cursor)[0] = Dice{Value: 0}

	// Setup player entities
	var player goke.Comp[Player]
	playerFactory := ecs.NewFactory(&player)

	playerFactory.Create(2)
	fc := &playerFactory.Cursor
	for playerFactory.Next() {
		players := player.Slice(fc)
		for i := range playerFactory.IDs {
			players[i] = Player{Bet: 0}
		}
	}

	// 3. Define Queries (for system filtering)
	vDice := ecs.NewQueryBuilder(&dice).Build()
	vPlayers := ecs.NewQueryBuilder(&player).Build()
	vWinners := ecs.NewQueryBuilder().Include(goke.Include[Winner]()).Build()

	diceCursor := &vDice.Cursor
	playerCursor := &vPlayers.Cursor

	// Query for reading the dice entity's value each turn
	diceQuery := ecs.NewQueryBuilder(&dice).Build()

	// 4. Register Systems

	// System A: Roll the dice
	rollSys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		vDice.All()
		for vDice.Next() {
			diceSlice := dice.Slice(diceCursor)
			for i := range vDice.Cursor.IDs {
				diceSlice[i].Value = rand.Intn(6) + 1
			}
		}
	})

	// System B: Players place their bets
	betSys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		vPlayers.All()
		for vPlayers.Next() {
			players := player.Slice(playerCursor)
			for i := range vPlayers.Cursor.IDs {
				players[i].Bet = rand.Intn(6) + 1
			}
		}
	})

	// System C: Judge the results
	judgeSys := ecs.RegSysFn(func(schedule *goke.CmdBuf, d time.Duration) {
		if gameFinished {
			return
		}
		turnCounter++

		diceQuery.Seek(diceEnt)
		diceComp := dice.At(&diceQuery.Cursor)
		fmt.Printf("🎲 Turn %d | Dice Result: %d\n", turnCounter, diceComp.Value)

		vPlayers.All()
		for vPlayers.Next() {
			players := player.Slice(playerCursor)
			for i, entityID := range vPlayers.Cursor.IDs {
				bet := players[i].Bet
				fmt.Printf("   Player %d bet: %d\n", entityID, bet)
				if bet == diceComp.Value {
					gameFinished = true
					// Defer the assignment of the Winner tag to the next Sync point
					goke.CmdBufAddComp(schedule, entityID, winnerID, Winner{})
				}
			}
		}
	})

	// System D: Display winners (Reactive System)
	displayWinnerSys := ecs.RegSysFn(func(cb *goke.CmdBuf, d time.Duration) {
		vWinners.All()
		for vWinners.Next() {
			for _, e := range vWinners.Cursor.IDs {
				fmt.Printf("🏆 VICTORY! Entity %d is marked as a Winner!\n", e)
			}
		}
	})

	// 5. Define Execution Plan
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
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
		ecs.Tick(16 * time.Millisecond)
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Simulation ended.")
}
