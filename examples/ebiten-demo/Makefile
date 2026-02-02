# Goke - Ebiten Integration Demo

.PHONY: all deps run tidy clean

all: run

# Ensure go.mod in this directory is fully populated
deps:
	go mod tidy

# Run the demo using the local directory as the root
run: deps
	go run .

tidy:
	go mod tidy

clean:
	go clean