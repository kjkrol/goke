package main

import (
	"log"

	"github.com/kjkrol/gokg/pkg/plane"
	"github.com/kjkrol/gokg/pkg/spatial"
)

type Grid struct {
	witdh, height uint32
	space         plane.Space2D[uint32]
	spatialIndex  *spatial.GridIndexManager
}

func NewGrid(witdh, height uint32, bucketCapacity int, opsBufferSize int) *Grid {
	config := spatial.GridIndexConfig{
		Resolution:       spatial.Size1024x1024,
		BucketResolution: spatial.Size32x32,
		BucketCapacity:   bucketCapacity,
		OpsBufferSize:    opsBufferSize,
	}

	space := plane.NewToroidal2D[uint32](witdh, height)
	spatialIndex, err := spatial.NewGridIndexManager(space, config)
	if err != nil {
		log.Fatalf("Failed to create bucket grid: %v", err)
	}

	return &Grid{
		witdh:        witdh,
		height:       height,
		space:        space,
		spatialIndex: spatialIndex,
	}

}
