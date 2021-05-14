package pm_logging

import (
	"testing"
)

func TestDefaultDecider(t *testing.T) {
	t.Parallel()

	if got := DefaultLogDecider(nil, nil); got != true {
		t.Errorf("DefaultLogDecider() = %v, want %v", got, true)
	}
}
