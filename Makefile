GO       = go
GOROOT  := $(shell go env GOROOT)

COMMIT_DATE := $(shell git log -1 --format=%cs)
COMMIT_HASH := $(shell git log -1 --format=%h)
DIRTY := $(shell git diff --quiet || echo "-dirty_$$(date +%H%M%S)")
RESULT_FILE := bench_results/bench_$(COMMIT_DATE)_$(COMMIT_HASH)$(DIRTY).txt

BENCH_COUNT ?= 4

.PHONY: setup
setup:
	go run internal/cmd/scripts/setup/setup_work.go
test:
	go test ./...
bench:
	go test -bench=. -benchmem ./...
bench-save:
	@mkdir -p bench_results
	go test -bench=. -benchmem -count=10 ./internal/bench/... > $(RESULT_FILE)
	@echo "Results saved into: $(RESULT_FILE)"

## bench-stats: Runs benchmarks multiple times and calculates P95/P99 latencies
bench-p95:
	@echo "Starting ultra-performance analysis (Count: $(BENCH_COUNT))..."
	@cd internal/cmd/scripts/bench && go run bench_parser.go $(BENCH_COUNT)
