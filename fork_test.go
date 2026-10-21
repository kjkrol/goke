package goke_test

import (
	"sync"
	"testing"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type forkPos struct{ X float64 }
type forkTag struct{ N int32 }

// forkWorld is a world of n entities with a forkPos at their index, every other one carrying a
// forkTag too, and a query over forkPos
func forkWorld(t *testing.T, n int) (*goke.Query, *goke.Comp[forkPos], []uid.UID64) {
	t.Helper()
	ecs := goke.New()
	pos := new(goke.Comp[forkPos])
	var q *goke.Query
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		var tag goke.Comp[forkTag]
		plain, tagged := si.NewFactory(pos), si.NewFactory(pos, &tag)
		for i := range n {
			f := plain
			if i%2 == 1 {
				f = tagged
			}
			f.Create(1)
			for f.Next() {
				pos.At(&f.Cursor).X = float64(len(ids))
				ids = append(ids, f.Cursor.IDs[0])
			}
		}
		q = si.NewQueryBuilder(pos).Build()
	}})
	return q, pos, ids
}

func count(q *goke.Query) int {
	n := 0
	for q.All(); q.Next(); {
		n += len(q.Cursor().IDs)
	}
	return n
}

// A fork walks what its query matches, and walks on its own: a Seek on one leaves the other's
// walk where it stood, and a Clear on one ends that one's walk only.
func TestQuery_Fork_WalksOnItsOwn(t *testing.T) {
	q, pos, ids := forkWorld(t, 6)
	f := q.Fork()
	if count(q) != 6 || count(f) != 6 {
		t.Fatalf("walked %d and %d entities, want 6 each", count(q), count(f))
	}

	q.All()
	if !q.Next() {
		t.Fatal("the query found nothing")
	}
	first := q.Cursor().IDs[0]
	for i, id := range ids {
		if !f.Seek(id) || pos.At(f.Cursor()).X != float64(i) {
			t.Fatalf("fork Seek(%v): want X %d", id, i)
		}
	}
	if q.Cursor().IDs[0] != first {
		t.Error("the fork's Seeks moved the query's walk")
	}
	n := 1
	for q.Next() {
		n++
	}
	if n != 2 { // a chunk per archetype
		t.Errorf("the query walked on through %d chunks, want 2", n)
	}

	f.All()
	f.Clear()
	if f.Next() {
		t.Error("the fork walks on after its Clear")
	}
	if count(q) != 6 || count(f) != 6 {
		t.Error("a Clear changed what the queries match")
	}
}

// Forks seek and walk on goroutines of their own, reading components none of them writes.
func TestQuery_Fork_SeeksOnManyGoroutines(t *testing.T) {
	q, pos, ids := forkWorld(t, 200)
	var wg sync.WaitGroup
	errs := make(chan string, 8)
	for g := range 8 {
		f := q.Fork()
		wg.Add(1)
		go func() {
			defer wg.Done()
			for round := range 20 {
				for i := (g + round) % len(ids); i < len(ids); i += 3 {
					if !f.SeekH(ids[i]) && !f.Seek(ids[i]) || pos.At(f.Cursor()).X != float64(i) {
						errs <- "a fork read another entity's component"
						return
					}
				}
				if count(f) != len(ids) {
					errs <- "a fork walked a different number of entities"
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
