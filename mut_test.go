package goke_test

import (
	"testing"
	"time"
	"unsafe"

	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type moved struct{ X, Y float64 }
type marked struct{ N uint32 }

// chunkVers is each chunk's Ver and its table's TableVer as a reader walking q sees them.
func chunkVers(t *testing.T, q *goke.Query) (map[unsafe.Pointer]uint32, uint32) {
	t.Helper()
	vers := map[unsafe.Pointer]uint32{}
	var tableVer uint32
	for q.All(); q.Next(); {
		s := q.ChunkSnapshot()
		if s.Ver != *q.Cursor().Ver {
			t.Fatalf("ChunkSnapshot.Ver %d, Cursor.Ver %d", s.Ver, *q.Cursor().Ver)
		}
		vers[s.ChunkPtr] = s.Ver
		tableVer = s.TableVer
	}
	return vers, tableVer
}

func changed(before, after map[unsafe.Pointer]uint32) int {
	n := 0
	for ptr, v := range after {
		if before[ptr] != v {
			n++
		}
	}
	return n
}

// Mut counts a change of every chunk it writes, Comp of none, and TableVer moves only with a
// removal or a relocation
func TestMut_CountsTheChunksItWrites(t *testing.T) {
	ecs := goke.New()
	var read goke.Comp[moved]
	var write goke.Mut[moved]
	var reader, writer *goke.Query
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&write)
		f.Create(3000)
		for f.Next() {
			for i := range write.Slice(&f.Cursor) {
				write.Slice(&f.Cursor)[i].X = float64(i)
			}
			ids = append(ids, f.IDs...)
		}
		reader = si.NewQueryBuilder(&read).Build()
		writer = si.NewQueryBuilder(&write).Build()
	}})
	start, tableVer := chunkVers(t, reader)
	if len(start) < 3 {
		t.Fatalf("3000 entities lie in %d chunks; the test needs several", len(start))
	}

	for reader.All(); reader.Next(); {
		_ = read.Slice(reader.Cursor())[0]
	}
	reader.Seek(ids[0])
	read.At(reader.Cursor()).Y = 1
	if now, _ := chunkVers(t, reader); changed(start, now) != 0 {
		t.Errorf("Comp counted a change of %d chunks", changed(start, now))
	}

	for writer.All(); writer.Next(); {
		write.Slice(writer.Cursor())[0].Y = 2
	}
	walked, now := chunkVers(t, reader)
	if changed(start, walked) != len(start) || now != tableVer {
		t.Errorf("a Mut walk changed %d of %d chunks, TableVer %d to %d",
			changed(start, walked), len(start), tableVer, now)
	}

	last := ids[len(ids)-1]
	writer.Seek(last)
	write.At(writer.Cursor()).Y = 3
	writer.SeekH(ids[0])
	write.At(writer.Cursor()).Y = 3
	for writer.Pick([]uid.UID64{last}); writer.Next(); {
		write.At(writer.Cursor()).Y = 4
	}
	sought, _ := chunkVers(t, reader)
	if changed(walked, sought) != 2 {
		t.Errorf("Seek, SeekH and Pick on two entities changed %d chunks, want 2",
			changed(walked, sought))
	}
	if write.Index() != 0 || len(goke.LoadComps(&write)) != 1 {
		t.Error("Mut is not bound as a Comp would be")
	}
}

// goke counts its own writes: an AddOne over a component the entity has, an append; TableVer
// moves with a removal
func TestChunkVer_CountsGokesOwnWrites(t *testing.T) {
	ecs := goke.New()
	movedID := ecs.RegComp[moved]()
	var read goke.Comp[moved]
	var reader *goke.Query
	var f *goke.Factory
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f = si.NewFactory(&read)
		f.Create(10)
		for f.Next() {
			ids = append(ids, f.IDs...)
		}
		reader = si.NewQueryBuilder(&read).Build()
	}})
	step := func(do func(cb *goke.CmdBuf)) {
		sys := ecs.RegSys(goke.SystemFn{OnUpdate: func(cb *goke.CmdBuf, _ time.Duration) {
			do(cb)
		}})
		ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
			ctx.Run(sys, d)
			_ = ctx.Sync()
		})
		ecs.Tick(time.Millisecond)
	}

	before, tableVer := chunkVers(t, reader)
	step(func(cb *goke.CmdBuf) { cb.AddOne(ids[0], movedID, moved{X: 9}) })
	after, now := chunkVers(t, reader)
	if changed(before, after) != 1 || now != tableVer {
		t.Errorf("an in-place AddOne changed %d chunks, TableVer %d to %d", changed(before, after),
			tableVer, now)
	}

	step(func(*goke.CmdBuf) {
		f.Create(1)
		for f.Next() {
		}
	})
	appended, now := chunkVers(t, reader)
	if changed(after, appended) != 1 || now != tableVer {
		t.Errorf("an append changed %d chunks, TableVer %d to %d", changed(after, appended),
			tableVer, now)
	}

	step(func(cb *goke.CmdBuf) { cb.RemoveOne(ids[0]) })
	if _, now := chunkVers(t, reader); now == tableVer {
		t.Error("a removal left TableVer as it was")
	}
}
