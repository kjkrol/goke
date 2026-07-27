package reg

import "testing"

func TestIsPowerOfTwo(t *testing.T) {
	cases := []struct {
		n    uint64
		want bool
	}{
		{0, false},
		{1, true},
		{2, true},
		{3, false},
		{4, true},
		{6, false},
		{64, true},
	}
	for _, c := range cases {
		if got := isPowerOfTwo(c.n); got != c.want {
			t.Errorf("isPowerOfTwo(%d) = %v, want %v", c.n, got, c.want)
		}
	}
}

func TestValidateConst_DoesNotPanicForRealHashSize(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("validateConst() panicked unexpectedly: %v", r)
		}
	}()
	validateConst()
}
