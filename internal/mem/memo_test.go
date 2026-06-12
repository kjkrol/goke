package mem

import (
	"testing"

	"github.com/kjkrol/goke/internal/core"
)

func TestMemo_LenTracking(t *testing.T) {
	compInfos := []core.ComponentInfo{
		{ID: 1, Size: 8, Align: 8},
	}

	var memo Memo
	memo.Init(compInfos)

	if memo.Len != 0 {
		t.Errorf("Expected initial Memo.Len to be 0, got %d", memo.Len)
	}

	memo.AllocSlot()
	memo.AllocSlot()
	memo.AllocSlot()

	if memo.Len != 3 {
		t.Errorf("Expected Memo.Len to be 3 after 3 allocations, got %d", memo.Len)
	}

	if memo.Pages[0].Len != 3 {
		t.Errorf("Expected page.Len to be 3, got %d", memo.Pages[0].Len)
	}

	memo.Clear()
	if memo.Len != 0 {
		t.Errorf("Expected Memo.Len to be 0 after Clear, got %d", memo.Len)
	}
}

func TestMemo_ResolveTail_Reserved(t *testing.T) {
	compInfos := []core.ComponentInfo{
		{ID: 1, Size: 8, Align: 8},
	}

	var memo Memo
	memo.Init(compInfos)

	memo.AddPages(4)

	if len(memo.Pages) != 5 {
		t.Fatalf("Expected 5 pages initially, got %d", len(memo.Pages))
	}

	memo.Reserved = 0
	tailIdx, _ := memo.ResolveTail()
	if tailIdx != 0 {
		t.Errorf("Expected tailIdx 0, got %d", tailIdx)
	}
	if len(memo.Pages) != 1 {
		t.Errorf("Expected pages to be truncated to 1, got %d", len(memo.Pages))
	}

	memo.AddPages(4)
	memo.Reserved = 2

	tailIdx, _ = memo.ResolveTail()

	if tailIdx != 0 {
		t.Errorf("Expected tailIdx 0 since no data exists, got %d", tailIdx)
	}
	if len(memo.Pages) != 3 {
		t.Errorf("Expected pages slice to be truncated to 3 (protecting reserved index 2), got %d", len(memo.Pages))
	}

	memo.AddPages(2)
	memo.Pages[4].Len = 1
	memo.Reserved = 2

	tailIdx, _ = memo.ResolveTail()

	if tailIdx != 4 {
		t.Errorf("Expected tailIdx 4, got %d", tailIdx)
	}
	if len(memo.Pages) != 5 {
		t.Errorf("Expected pages to remain 5, got %d", len(memo.Pages))
	}
}
