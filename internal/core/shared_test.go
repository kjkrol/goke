package core_test

// shared_test.go contains common component types used across
// multiple test suites to avoid duplication.

type position struct {
	x, y float64
}

type velocity struct {
	vx, vy float64
}

type rotation struct {
	angle float32
}
