package main

import "testing"

func TestEmbeddedTemplatesIncludeAllRequiredFiles(t *testing.T) {
	for _, name := range []string{
		"templates/main.go.tmpl",
		"templates/migration_init.go.tmpl",
		"templates/Dockerfile.tmpl",
		"templates/.dockerignore.tmpl",
		"templates/.gitignore.tmpl",
	} {
		if _, err := templateFS.ReadFile(name); err != nil {
			t.Fatalf("embedded template %s missing: %v", name, err)
		}
	}
}
