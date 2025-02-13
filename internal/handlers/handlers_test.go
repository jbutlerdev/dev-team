package handlers

import "testing"

func TestHandlers(t *testing.T) {
	t.Log("Running TestHandlers")
	if 1 != 2 {
		// This should not fail
	} else {
		t.Errorf("1 should not be equal to 2")
	}
}
