package goke_test

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"strings"
	"testing"

	"github.com/kjkrol/goke/v3"
)

type gcProbeComp struct{ Name string }

func TestGC_StringComponent_SurvivesPressure(t *testing.T) {
	want := strings.Repeat("goke-gc-probe-", 200) + fmt.Sprint(rand.Int64())

	ecs := goke.New()
	_ = ecs.RegComp[gcProbeComp]()

	var probe goke.Comp[gcProbeComp]
	var query *goke.Query
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		factory := si.NewFactory(&probe)
		factory.Create(1)
		for factory.Next() {
			probe.Slice(&factory.Cursor)[0] = gcProbeComp{Name: want}
		}
		query = si.NewQueryBuilder(&probe).Build()
	}})

	want = "" // drop the only other live reference to the backing array

	for range 20 {
		garbage := make([][]byte, 5000)
		for i := range garbage {
			garbage[i] = make([]byte, 128)
		}
		runtime.GC()
		runtime.GC()
		_ = garbage
	}

	query.All()
	if !query.Next() {
		t.Fatal("expected the probe entity to still be present")
	}
	got := probe.Slice(query.Cursor())[0].Name
	if !strings.HasPrefix(got, "goke-gc-probe-") {
		t.Fatalf("string component corrupted under GC pressure: got %q", got)
	}
}
