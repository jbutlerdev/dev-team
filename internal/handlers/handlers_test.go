package handlers

import (
	"html/template"
	"testing"
)

func TestHandlers(t *testing.T) {
	t.Run("Template Initialization", func(t *testing.T) {
		templates = nil // Reset templates to nil for testing
		tmpl := template.Must(template.New("test").Parse("<html></html>"))
		InitTemplates(tmpl)
		if templates == nil {
			t.Errorf("Templates should not be nil after initialization")
		}
	})

	// Test case: Check if 1 + 1 = 2
	t.Run("TestAddition", func(t *testing.T) {
		sum := 1 + 1
		if sum != 2 {
			t.Errorf("1 + 1 should be 2, but got %d", sum)
		}
	})

	// Test case: Check if a string is not empty
	t.Run("TestStringNotEmpty", func(t *testing.T) {
		str := "hello"
		if str == "" {
			t.Errorf("String should not be empty")
		}
	})

	// Test case: Check if 2 * 2 = 4
	t.Run("TestMultiplication", func(t *testing.T) {
		product := 2 * 2
		if product != 4 {
			t.Errorf("2 * 2 should be 4, but got %d", product)
		}
	})

	// Test case: Check if a boolean is true
	t.Run("TestBooleanTrue", func(t *testing.T) {
		value := true
		if !value {
			t.Errorf("Boolean should be true")
		}
	})
}
