package handlers

import "testing"

func TestHandlers(t *testing.T) {
	// TODO: Add tests here
	t.Run("Simple Test", func(t *testing.T) {
		x := 1
		if true != false {
			if x != 1 {
				t.Errorf("x is not equal to 1")
			}
		}
	})
}
