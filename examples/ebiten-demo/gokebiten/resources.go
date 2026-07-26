package gokebiten

import (
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/control"
	"github.com/kjkrol/goke/examples/ebiten-demo/gokebiten/spatial"
)

// Resources bundles what Game needs (GameProps, input, world config, and
// built-in TPS/EntityCount telemetry) with two game-owned shapes: S for
// control state, T for display telemetry.
type Resources[S, T any] struct {
	gameProps   *GameProps
	inputs      control.InputEvents
	spaceConfig spatial.Config
	measuredTPS int

	state     S
	telemetry T
}

var _ resources = (*Resources[struct{}, struct{}])(nil)

func NewResources[S, T any](gameProps *GameProps, spaceConfig spatial.Config, state S, telemetry T) *Resources[S, T] {
	return &Resources[S, T]{gameProps: gameProps, spaceConfig: spaceConfig, state: state, telemetry: telemetry}
}

func (r *Resources[S, T]) State() *S     { return &r.state }
func (r *Resources[S, T]) Telemetry() *T { return &r.telemetry }
func (r *Resources[S, T]) TPS() *int     { return &r.measuredTPS }

func (r *Resources[S, T]) GetGameProps() *GameProps             { return r.gameProps }
func (r *Resources[S, T]) GetSpaceConfig() spatial.Config       { return r.spaceConfig }
func (r *Resources[S, T]) GetInputEvents() *control.InputEvents { return &r.inputs }

// Resettable lets Telemetry hook into each stats interval — if T implements
// it, Reset is called right after TPS is refreshed.
type Resettable interface{ Reset() }

func (r *Resources[S, T]) Reset() {
	if resettable, ok := any(&r.telemetry).(Resettable); ok {
		resettable.Reset()
	}
}
