package kinematics

// Telemetry is how many entities Populate/PopulateStatic have created —
// own this as a field in your own Telemetry type and pass a pointer to
// Game.Populate/PopulateStatic.
type Telemetry struct {
	DynamicCount int
	StaticCount  int
}
