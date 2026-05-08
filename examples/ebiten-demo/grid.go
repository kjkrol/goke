package main

import (
	"log"

	"github.com/kjkrol/gokg/pkg/plane"
	"github.com/kjkrol/gokg/pkg/spatial"
)

type Grid struct {
	space        plane.Space2D[uint32]
	spatialIndex *spatial.GridIndexManager
	gridConfig   spatial.GridIndexConfig
}

func NewGrid(config spatial.GridIndexConfig) *Grid {
	witdh := config.Resolution.Side()
	height := config.Resolution.Side()

	log.Printf("w=%d, h=%d, c=%v", witdh, height, config)

	space := plane.NewToroidal2D[uint32](witdh, height)
	spatialIndex, err := spatial.NewGridIndexManager(space, config)
	if err != nil {
		log.Fatalf("Failed to create bucket grid: %v", err)
	}

	return &Grid{
		space:        space,
		spatialIndex: spatialIndex,
		gridConfig:   config,
	}

}
