package arch

import (
	"testing"

	"github.com/kjkrol/uid"

	"github.com/kjkrol/goke/internal/comp"
)

func TestGraph_InitialState(t *testing.T) {
	var g Graph

	if g.CountNextEdges() != 0 {
		t.Errorf("expected 0 next edges, got %d", g.CountNextEdges())
	}
	if g.CountPrevEdges() != 0 {
		t.Errorf("expected 0 prev edges, got %d", g.CountPrevEdges())
	}
}

func TestGraph_CountNextEdges(t *testing.T) {
	var g Graph

	g.edgesNext[comp.ID(0)] = ID(1)
	g.edgesNext[comp.ID(3)] = ID(2)

	if got := g.CountNextEdges(); got != 2 {
		t.Errorf("expected 2 next edges, got %d", got)
	}
}

func TestGraph_CountPrevEdges(t *testing.T) {
	var g Graph

	g.edgesPrev[comp.ID(1)] = ID(3)

	if got := g.CountPrevEdges(); got != 1 {
		t.Errorf("expected 1 prev edge, got %d", got)
	}
}

func TestGraph_Reset(t *testing.T) {
	var g Graph

	g.edgesNext[comp.ID(0)] = ID(1)
	g.edgesNext[comp.ID(1)] = ID(2)
	g.edgesPrev[comp.ID(0)] = ID(3)

	g.Reset()

	if g.CountNextEdges() != 0 {
		t.Errorf("expected 0 next edges after Reset, got %d", g.CountNextEdges())
	}
	if g.CountPrevEdges() != 0 {
		t.Errorf("expected 0 prev edges after Reset, got %d", g.CountPrevEdges())
	}
}

func TestGraph_LinkNextArch(t *testing.T) {
	env := newTestEnv()
	posInfo, _ := testMetas()

	env.addEntity(uid.UID64(1), RootID)
	env.upsertComp(uid.UID64(1), posInfo)

	root := &env.catalog.Archetypes[RootID]
	posArchID := root.graph.edgesNext[posInfo.ID]
	if posArchID == NullID {
		t.Fatal("expected next edge for position component")
	}

	posArch := &env.catalog.Archetypes[posArchID]
	if posArch.graph.edgesPrev[posInfo.ID] != RootID {
		t.Error("expected bidirectional prev edge back to root")
	}
}

func TestGraph_MultipleNextEdges(t *testing.T) {
	env := newTestEnv()
	posInfo, velInfo := testMetas()

	env.addEntity(uid.UID64(1), RootID)
	env.addEntity(uid.UID64(2), RootID)
	env.upsertComp(uid.UID64(1), posInfo)
	env.upsertComp(uid.UID64(2), velInfo)

	root := &env.catalog.Archetypes[RootID]
	if got := root.graph.CountNextEdges(); got != 2 {
		t.Errorf("expected 2 next edges from root, got %d", got)
	}

	posArchID := root.graph.edgesNext[posInfo.ID]
	velArchID := root.graph.edgesNext[velInfo.ID]
	if posArchID == velArchID {
		t.Error("position and velocity should lead to different archetypes")
	}
}
