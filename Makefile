GO       = go
GOROOT  := $(shell go env GOROOT)

DEMO ?= cmd/ecs-demo
DEMO_DIR := $(patsubst %/,%,$(dir $(DEMO)))
GO_DEMO := $(if $(filter %.go,$(DEMO)),$(DEMO_DIR),$(DEMO))
GO_DEMO_PKG := ./$(GO_DEMO)

# ------------------- RUN -------------------

run-demo:
	$(GO) run $(GO_DEMO_PKG)

# ------------------- CLEAN -----------------

clear:
	rm -rf bin/
