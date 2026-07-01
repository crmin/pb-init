package initcli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMainTemplateUsesModulePathMigrationDirJSVMAndAutoMigration(t *testing.T) {
	dir := renderFixture(t, Config{
		MigrationDir:  "internal/migrations",
		AutoMigration: true,
		JSVM:          true,
	})

	mainGo := readFile(t, dir, "main.go")
	for _, want := range []string{
		`_ "example.com/app/internal/migrations"`,
		`"github.com/pocketbase/pocketbase/plugins/jsvm"`,
		`Automigrate: true`,
		`jsvm.MustRegister(app, jsvm.Config{})`,
	} {
		if !strings.Contains(mainGo, want) {
			t.Fatalf("main.go missing %q:\n%s", want, mainGo)
		}
	}
}

func TestRenderMigrationInitUsesPackageNameFromNestedDir(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: "internal/migrations"})

	got := strings.TrimSpace(readFile(t, dir, "internal/migrations/init.go"))
	if got != "package migrations" {
		t.Fatalf("unexpected migration init content: %q", got)
	}
}

func TestRenderCreatesNestedMigrationDirectory(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: "internal/migrations"})

	info, err := os.Stat(filepath.Join(dir, "internal", "migrations"))
	if err != nil {
		t.Fatalf("expected migration directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected migration path to be directory")
	}
}

func TestRenderDockerFilesUseCgoAndBinaryName(t *testing.T) {
	dir := renderFixture(t, Config{
		MigrationDir: defaultMigrationDir,
		Docker:       true,
		CgoEnabled:   true,
	})

	dockerfile := readFile(t, dir, "Dockerfile")
	if !strings.Contains(dockerfile, "RUN CGO_ENABLED=1 go build -o pocketbase .") {
		t.Fatalf("Dockerfile missing CGO setting:\n%s", dockerfile)
	}

	dockerignore := readFile(t, dir, ".dockerignore")
	if !strings.Contains(dockerignore, "\napp\n") {
		t.Fatalf(".dockerignore missing binary name:\n%s", dockerignore)
	}
}

func TestRenderCreatesJSVMAssetDirectoriesWhenJSVMEnabled(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir, JSVM: true})

	for _, name := range []string{"pb_migrations", "pb_hooks"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("expected %s directory: %v", name, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be directory", name)
		}
	}
}

func TestRenderSkipsJSVMAssetDirectoriesWhenJSVMDisabled(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir})

	for _, name := range []string{"pb_migrations", "pb_hooks"} {
		_, err := os.Stat(filepath.Join(dir, name))
		if !os.IsNotExist(err) {
			t.Fatalf("expected %s to be absent, err=%v", name, err)
		}
	}
}

func TestRenderDockerfileCopiesJSVMAssetDirectoriesWhenJSVMEnabled(t *testing.T) {
	dir := renderFixture(t, Config{
		MigrationDir: defaultMigrationDir,
		Docker:       true,
		JSVM:         true,
	})

	dockerfile := readFile(t, dir, "Dockerfile")
	for _, want := range []string{
		"RUN mkdir -p pb_migrations pb_hooks",
		"COPY --from=builder /go/src/app/pb_migrations /pb_migrations",
		"COPY --from=builder /go/src/app/pb_hooks /pb_hooks",
	} {
		if !strings.Contains(dockerfile, want) {
			t.Fatalf("Dockerfile missing %q:\n%s", want, dockerfile)
		}
	}
}

func TestRenderDockerfileOmitsJSVMAssetDirectoriesWhenJSVMDisabled(t *testing.T) {
	dir := renderFixture(t, Config{
		MigrationDir: defaultMigrationDir,
		Docker:       true,
	})

	dockerfile := readFile(t, dir, "Dockerfile")
	for _, notWant := range []string{"pb_migrations", "pb_hooks"} {
		if strings.Contains(dockerfile, notWant) {
			t.Fatalf("Dockerfile should not contain %q:\n%s", notWant, dockerfile)
		}
	}
}

func TestRenderBinaryNamePocketBaseIsOmitted(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{MigrationDir: defaultMigrationDir, Docker: true}
	err := RenderProject(Project{Dir: dir, ModulePath: "example.com/pocketbase"}, cfg, testTemplateFS())
	if err != nil {
		t.Fatalf("RenderProject returned error: %v", err)
	}

	gitignore := readFile(t, dir, ".gitignore")
	if strings.Count(gitignore, "\npocketbase\n") != 1 {
		t.Fatalf("expected no duplicate pocketbase binary name:\n%s", gitignore)
	}
}

func TestRenderAlwaysWritesGitignore(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir})

	gitignore := readFile(t, dir, ".gitignore")
	if !strings.Contains(gitignore, "pb_data/*") {
		t.Fatalf(".gitignore missing pb_data entry:\n%s", gitignore)
	}
}

func renderFixture(t *testing.T, cfg Config) string {
	t.Helper()
	dir := t.TempDir()
	if cfg.MigrationDir == "" {
		cfg.MigrationDir = defaultMigrationDir
	}

	err := RenderProject(Project{Dir: dir, ModulePath: "example.com/app"}, cfg, testTemplateFS())
	if err != nil {
		t.Fatalf("RenderProject returned error: %v", err)
	}
	return dir
}

func testTemplateFS() fs.FS {
	return os.DirFS(filepath.Join("..", ".."))
}

func readFile(t *testing.T, dir string, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", name, err)
	}
	return string(content)
}
