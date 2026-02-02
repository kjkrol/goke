# Goke - Ebiten Integration Demo

This is a minimal demonstration of the **goke** engine integrated with **Ebitengine**. 

It showcases the `Space2D[uint32]` plane handling real-time object management and collision boundaries.

## Prerequisites
- Go 1.21+
- [Ebitengine dependencies](https://ebitengine.org/en/documents/install.html) (C compiler and system libraries)

## Setup & Execution

You don't need to install packages manually. The `Makefile` handles it via `go mod tidy`:

```bash
# To install dependencies and start the demo
make run
```