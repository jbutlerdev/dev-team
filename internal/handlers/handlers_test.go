package handlers

import (
	"html/template"
	"testing"
)

func TestHandlers(t *testing.T) {
	t.Log("Running TestHandlers")
	// Create a mock template
	mockTemplate := template.Must(template.New("test").Parse("<h1>Test</h1>"))

	// Call InitTemplates with the mock template
	InitTemplates(mockTemplate)

	// Verify that the templates variable is no longer nil
	if templates == nil {
		t.Errorf("templates variable should not be nil after calling InitTemplates")
	}
}
