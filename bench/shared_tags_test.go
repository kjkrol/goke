package bench_test

import (
	"github.com/kjkrol/goke/v3"
)

type Tag1 struct{}
type Tag2 struct{}
type Tag3 struct{}
type Tag4 struct{}
type Tag5 struct{}
type Tag6 struct{}
type Tag7 struct{}
type Tag8 struct{}
type Tag9 struct{}
type Tag10 struct{}

func setupTagECS() *goke.ECS {
	ecs := setupECS()
	_ = ecs.RegComp[Tag1]()
	_ = ecs.RegComp[Tag2]()
	_ = ecs.RegComp[Tag3]()
	_ = ecs.RegComp[Tag4]()
	_ = ecs.RegComp[Tag5]()
	_ = ecs.RegComp[Tag6]()
	_ = ecs.RegComp[Tag7]()
	_ = ecs.RegComp[Tag8]()
	_ = ecs.RegComp[Tag9]()
	_ = ecs.RegComp[Tag10]()
	return ecs
}
