package ecs_test

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

// Any other shared test utilities can go here, for example:
type complexComponent struct {
	Active bool
	Layer  int32
	Name   [16]byte
}

type testComponent struct {
	ID     int64
	Value  float64
	Active bool
}
