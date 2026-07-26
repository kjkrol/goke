package stats

// Stats is the aggregate telemetry this strategy maintains — own it as a
// field in your Telemetry type and hand a pointer to NewHandler.
type Stats struct {
	Counter int
}
