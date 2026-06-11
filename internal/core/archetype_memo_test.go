package core

import (
	"testing"
)

func TestMemo_LenTracking(t *testing.T) {
	compInfos := []ComponentInfo{
		{ID: 1, Size: 8, Align: 8},
	}

	var memo Memo
	memo.Init(compInfos)

	if memo.Len != 0 {
		t.Errorf("Expected initial Memo.Len to be 0, got %d", memo.Len)
	}

	// Allocate 3 slots
	memo.AllocSlot()
	memo.AllocSlot()
	memo.AllocSlot()

	if memo.Len != 3 {
		t.Errorf("Expected Memo.Len to be 3 after 3 allocations, got %d", memo.Len)
	}

	// Ensure the page length matches the global length for a single page
	if memo.Pages[0].Len != 3 {
		t.Errorf("Expected page.Len to be 3, got %d", memo.Pages[0].Len)
	}

	// Clear should reset the global length tracking
	memo.Clear()
	if memo.Len != 0 {
		t.Errorf("Expected Memo.Len to be 0 after Clear, got %d", memo.Len)
	}
}

func TestMemo_ResolveTail_Reserved(t *testing.T) {
	compInfos := []ComponentInfo{
		{ID: 1, Size: 8, Align: 8},
	}

	var memo Memo
	memo.Init(compInfos)

	// Artificially add pages (they start with Len == 0)
	memo.AddPages(4) // Initial 1 + 4 = 5 pages total (indices 0 to 4)

	if len(memo.Pages) != 5 {
		t.Fatalf("Expected 5 pages initially, got %d", len(memo.Pages))
	}

	// Case 1: No reserved pages, should truncate all empty trailing pages down to index 0
	memo.Reserved = 0
	tailIdx, _ := memo.ResolveTail()
	if tailIdx != 0 {
		t.Errorf("Expected tailIdx 0, got %d", tailIdx)
	}
	if len(memo.Pages) != 1 {
		t.Errorf("Expected pages to be truncated to 1, got %d", len(memo.Pages))
	}

	// Case 2: Reserved page prevents truncation of the slice
	memo.AddPages(4) // Back to 5 empty pages
	memo.Reserved = 2

	tailIdx, _ = memo.ResolveTail()

	// tailIdx should still be 0 (because the actual last page with data is 0)
	if tailIdx != 0 {
		t.Errorf("Expected tailIdx 0 since no data exists, got %d", tailIdx)
	}
	// However, the physical slice length must be 3 (indices 0, 1, 2) to protect the Reserved index
	if len(memo.Pages) != 3 {
		t.Errorf("Expected pages slice to be truncated to 3 (protecting reserved index 2), got %d", len(memo.Pages))
	}

	// Case 3: Data exists beyond the reserved index
	memo.AddPages(2)      // 3 + 2 = 5 pages
	memo.Pages[4].Len = 1 // Add fake data on the last page
	memo.Reserved = 2

	tailIdx, _ = memo.ResolveTail()

	// Should not truncate anything because the very last page has data
	if tailIdx != 4 {
		t.Errorf("Expected tailIdx 4, got %d", tailIdx)
	}
	if len(memo.Pages) != 5 {
		t.Errorf("Expected pages to remain 5, got %d", len(memo.Pages))
	}
}
