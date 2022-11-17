package generators

import "testing"

func TestRemaining(t *testing.T) {
	remaining := []int{1, 2, 3, 4, 5}
	for len(remaining) > 0 {
		// Pop the next element from the list
		next := remaining[0]
		remaining[0] = remaining[len(remaining)-1]
		remaining = remaining[:len(remaining)-1]
		t.Logf("next = %d\n", next)
		t.Logf("remaining = %#v\n", remaining)
	}
}
