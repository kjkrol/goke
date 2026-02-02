module github.com/kjkrol/goke/examples/ebiten-balls

go 1.25.3

// v0.0.0 is a placeholder; the 'replace' directive ensures the local source is used
require github.com/kjkrol/goke v1.0.0

require (
	github.com/hajimehoshi/ebiten/v2 v2.9.8
	github.com/kjkrol/gokg v1.2.4
)

require (
	github.com/ebitengine/gomobile v0.0.0-20250923094054-ea854a63cce1 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/purego v0.9.1 // indirect
	github.com/jezek/xgb v1.3.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
)

replace github.com/kjkrol/goke => ../../
