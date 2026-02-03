GO       = go
GOROOT  := $(shell go env GOROOT)
.PHONY: setup
setup:
	go run scripts/setup_work.go
