package main

import (
	"log"
	"math"

	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten"
	"github.com/kjkrol/gokg"
	"github.com/kjkrol/gokg/spatial"
)

// --- Configuration ---
const (
	TPS          = 60 * 2
	ScreenWidth  = 1024
	ScreenHeight = 1024
	RectSize     = 20 // rozmiar boku sprite'a w pikselach — bazowa jednostka dla adaptiveCapacity()
	FillPercent  = 20 // procent powierzchni świata wypełniony sprite'ami (0–100)
)

// EntityCount wyliczany z FillPercent — ile sprite'ów RectSize×RectSize potrzeba
// żeby osiągnąć żądany procent wypełnienia powierzchni ScreenWidth×ScreenHeight.
var EntityCount = int(math.Floor(FillPercent / 100.0 * float64(ScreenWidth*ScreenHeight) / float64(RectSize*RectSize)))

type Statistics struct {
	collisionCounter          int
	entityCounter             int
	measuredTPS               int
	meeasuredCollisionCounter int
}

type Resources struct {
	gameProps *gokebiten.GameProps
	space     *gokg.Space
	rectSize  int
	inputs    gokebiten.InputEvents
	Statistics
	gamePause bool
}

var _ gokebiten.Resources = (*Resources)(nil)

// adaptiveCapacity dobiera liczbę sprite'ów na bok bucketa na podstawie gęstości
// obiektów w świecie. Im wyższa gęstość, tym mniejszy bucket (rozprasza obiekty
// po większej liczbie bucketów → mniejsze listy w Query).
//
// Heurystyka: capacity = round(1 / sqrt(density)), clamped do [2, 8].
// Intuicja: jeśli sprite'y wypełniają 1/N² powierzchni, są przeciętnie oddalone
// o ~N rozmiarów sprite'a → bucket o boku N*RectSize naturalnie grupuje region.
//
// UWAGA: spatial.ResolutionFrom zaokrągla do potęgi 2, więc małe zmiany capacity
// mogą nie wpłynąć na faktyczny bucket size (np. capacity=4,5,6 dla RectSize=5
// dają to samo bucket=32×32).
func adaptiveCapacity() uint32 {
	const minCapacity, maxCapacity = 2.0, 8.0

	worldArea := uint64(ScreenWidth) * uint64(ScreenHeight)
	spriteArea := uint64(RectSize) * uint64(RectSize)
	density := float64(uint64(EntityCount)*spriteArea) / float64(worldArea)

	raw := math.Round(1.0 / math.Sqrt(density))
	clamped := math.Max(minCapacity, math.Min(maxCapacity, raw))
	return uint32(clamped)
}

func NewResources() *Resources {
	capacity := adaptiveCapacity()
	bucketResolution := spatial.ResolutionFrom(RectSize * capacity)

	log.Printf("[spatial] entities=%d, density=%.2f%%, capacity=%d → bucket=%dx%d, bucketCap=%d, opsBuffer=%d",
		EntityCount,
		float64(uint64(EntityCount)*uint64(RectSize)*uint64(RectSize))/float64(uint64(ScreenWidth)*uint64(ScreenHeight))*100,
		capacity,
		bucketResolution.Side(), bucketResolution.Side(),
		capacity*capacity,
		EntityCount*8)

	space, _ := gokg.NewSpace(gokg.Config{
		Width:          ScreenWidth,
		Height:         ScreenHeight,
		Toroidal:       true,
		BucketSize:     bucketResolution,
		BucketCapacity: int(capacity * capacity),
		OpsBufferSize:  EntityCount * 8,
	})
	return &Resources{
		gameProps: &gokebiten.GameProps{
			Title:        "GOKe + GOKg + Ebiten Integration",
			ScreenWidth:  ScreenWidth,
			ScreenHeight: ScreenHeight,
			TargetTPS:    TPS,
		},
		space:    space,
		rectSize: RectSize,
		Statistics: Statistics{
			entityCounter: EntityCount,
		},
	}
}

func (r *Resources) GetGameProps() *gokebiten.GameProps {
	return r.gameProps
}

func (r *Resources) Reset() {
	r.collisionCounter = 0
}

func (r *Resources) Refresh(tick int) {
	r.measuredTPS = tick
	r.meeasuredCollisionCounter = r.collisionCounter
}

func (r *Resources) GetInputEvents() *gokebiten.InputEvents { return &r.inputs }
