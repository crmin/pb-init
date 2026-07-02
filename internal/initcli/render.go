package initcli

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

const (
	templateMain         = "templates/main.go.tmpl"
	templateMigration    = "templates/migration_init.go.tmpl"
	templateDockerfile   = "templates/Dockerfile.tmpl"
	templateDockerignore = "templates/.dockerignore.tmpl"
	templateGitignore    = "templates/.gitignore.tmpl"
	templateJustfile     = "templates/justfile.tmpl"
)

// RenderProject writes the generated PocketBase project files.
func RenderProject(project Project, cfg Config, templates fs.FS) error {
	if templates == nil {
		return fmt.Errorf("template filesystem is required")
	}

	data := templateData(project, cfg)

	if err := renderTemplateFile(templates, templateMain, filepath.Join(project.Dir, "main.go"), data, true); err != nil {
		return err
	}

	migrationDir := filepath.Join(project.Dir, filepath.FromSlash(cfg.MigrationDir))
	if err := os.MkdirAll(migrationDir, 0o755); err != nil {
		return err
	}
	if err := renderTemplateFile(templates, templateMigration, filepath.Join(migrationDir, "init.go"), data, true); err != nil {
		return err
	}

	if cfg.JSVM {
		for _, name := range []string{"pb_migrations", "pb_hooks"} {
			if err := os.MkdirAll(filepath.Join(project.Dir, name), 0o755); err != nil {
				return err
			}
		}
	}

	if cfg.Docker {
		if err := renderTemplateFile(templates, templateDockerfile, filepath.Join(project.Dir, "Dockerfile"), data, false); err != nil {
			return err
		}
		if err := renderTemplateFile(templates, templateDockerignore, filepath.Join(project.Dir, ".dockerignore"), data, false); err != nil {
			return err
		}
	}

	if cfg.Just {
		if err := renderTemplateFile(templates, templateJustfile, filepath.Join(project.Dir, "justfile"), data, false); err != nil {
			return err
		}
	}

	return renderTemplateFile(templates, templateGitignore, filepath.Join(project.Dir, ".gitignore"), data, false)
}

type renderData struct {
	ModulePath       string
	MigrationDir     string
	MigrationPackage string
	JSVMImport       bool
	JSVMAssets       bool
	AutoMigration    string
	CgoEnabled       string
	BinaryName       string
	Justfile         bool
}

func templateData(project Project, cfg Config) renderData {
	return renderData{
		ModulePath:       project.ModulePath,
		MigrationDir:     cfg.MigrationDir,
		MigrationPackage: migrationPackageName(cfg.MigrationDir),
		JSVMImport:       cfg.JSVM,
		JSVMAssets:       cfg.JSVM,
		AutoMigration:    boolLiteral(cfg.AutoMigration),
		CgoEnabled:       cgoValue(cfg.CgoEnabled),
		BinaryName:       binaryName(project.ModulePath),
		Justfile:         cfg.Just,
	}
}

func renderTemplateFile(templates fs.FS, templateName string, target string, data renderData, goFile bool) error {
	content, err := fs.ReadFile(templates, templateName)
	if err != nil {
		return err
	}

	tmpl, err := template.New(path.Base(templateName)).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	output := buf.Bytes()
	if goFile {
		output, err = format.Source(output)
		if err != nil {
			return fmt.Errorf("format %s: %w", templateName, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	return os.WriteFile(target, output, 0o644)
}

func migrationPackageName(migrationDir string) string {
	base := path.Base(filepath.ToSlash(migrationDir))
	var b strings.Builder

	for i, r := range base {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			b.WriteRune(r)
			continue
		}
		if i == 0 && unicode.IsDigit(r) {
			b.WriteByte('_')
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}

	result := strings.Trim(b.String(), "_")
	if result == "" {
		return "_"
	}
	return result
}

func boolLiteral(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func cgoValue(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}

func binaryName(modulePath string) string {
	name := path.Base(modulePath)
	if name == "pocketbase" {
		return ""
	}
	return name
}
