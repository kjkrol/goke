GO       = go
GOROOT  := $(shell go env GOROOT)

BENCH_COUNT ?= 10

.PHONY: setup
setup:
	go run internal/cmd/scripts/setup/setup_work.go
test:
	go test ./...
bench:
	go test -bench=. -benchmem ./...
## bench-stats: Runs benchmarks multiple times and calculates P95/P99 latencies
bench-p95:
	@echo "Starting ultra-performance analysis (Count: $(BENCH_COUNT))..."
	@cd internal/cmd/scripts/bench && go run bench_parser.go $(BENCH_COUNT)
