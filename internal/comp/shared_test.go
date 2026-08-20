package comp_test

// shared_test.go contains common component types used across
// multiple test suites to avoid duplication.

type position struct {
	X, Y float64
}

type velocity struct {
	VX, VY float64
}

type rotation struct {
	Angle float32
}
