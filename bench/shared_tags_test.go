package bench_test

import (
	"github.com/kjkrol/goke/v2"
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
	_ = goke.RegComp[Tag1](ecs)
	_ = goke.RegComp[Tag2](ecs)
	_ = goke.RegComp[Tag3](ecs)
	_ = goke.RegComp[Tag4](ecs)
	_ = goke.RegComp[Tag5](ecs)
	_ = goke.RegComp[Tag6](ecs)
	_ = goke.RegComp[Tag7](ecs)
	_ = goke.RegComp[Tag8](ecs)
	_ = goke.RegComp[Tag9](ecs)
	_ = goke.RegComp[Tag10](ecs)
	return ecs
}
