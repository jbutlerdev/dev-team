package handlers

import (
	"fmt"
	"html/template"
)

var templates *template.Template

func InitTemplates(t *template.Template) {
	templates = t
	fmt.Println("Test Issue #0 has been addressed.")
}
