package initcli

import (
	"io/fs"
	"os"
	"os/exec"
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

func TestRenderWritesJustfileWhenJustEnabled(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir, Just: true})

	justfile := readFile(t, dir, "justfile")
	for _, want := range []string{
		"set positional-arguments := true",
		"[private]\ndefault:",
		"serve *args:",
		"migrate *args:",
		"snapshot *args:",
		"upgrade version=\"\":",
		"The following files will be deleted. Continue? (Y/n): ",
	} {
		if !strings.Contains(justfile, want) {
			t.Fatalf("justfile missing %q:\n%s", want, justfile)
		}
	}
}

func TestRenderJustfileUsesGoRunForMigrationCommands(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir, Just: true})

	justfile := readFile(t, dir, "justfile")
	for _, want := range []string{
		`go run . migrate collections "$@"`,
		`printf 'y\n' | go run . migrate collections`,
		`printf 'y\n' | go run . migrate collections "${migrate_args[@]}"`,
	} {
		if !strings.Contains(justfile, want) {
			t.Fatalf("justfile missing %q:\n%s", want, justfile)
		}
	}
	if strings.Contains(justfile, "./pocketbase migrate collections") {
		t.Fatalf("justfile should not depend on a prebuilt ./pocketbase binary:\n%s", justfile)
	}
}

func TestRenderJustfileUsesConfiguredMigrationDirInSnapshot(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: "internal/migrations", Just: true})

	justfile := readFile(t, dir, "justfile")
	for _, want := range []string{
		`migration_dir="internal/migrations"`,
		`if [[ ! -d "$migration_dir" ]]; then`,
		`find "$migration_dir" -maxdepth 1 -type f -name '*.go' -print | sort`,
	} {
		if !strings.Contains(justfile, want) {
			t.Fatalf("justfile missing %q:\n%s", want, justfile)
		}
	}
	for _, notWant := range []string{
		"[[ ! -d migrations ]]",
		"find migrations -maxdepth 1",
	} {
		if strings.Contains(justfile, notWant) {
			t.Fatalf("justfile should not contain hardcoded migration dir %q:\n%s", notWant, justfile)
		}
	}
}

func TestRenderSkipsJustfileWhenJustDisabled(t *testing.T) {
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir})

	_, err := os.Stat(filepath.Join(dir, "justfile"))
	if !os.IsNotExist(err) {
		t.Fatalf("expected justfile to be absent, err=%v", err)
	}
}

func TestRenderDockerignoreIncludesJustfileOnlyWhenJustEnabled(t *testing.T) {
	withJust := renderFixture(t, Config{
		MigrationDir: defaultMigrationDir,
		Docker:       true,
		Just:         true,
	})
	dockerignore := readFile(t, withJust, ".dockerignore")
	if !strings.Contains(dockerignore, "\njustfile\n") {
		t.Fatalf(".dockerignore missing justfile:\n%s", dockerignore)
	}

	withoutJust := renderFixture(t, Config{
		MigrationDir: defaultMigrationDir,
		Docker:       true,
	})
	dockerignore = readFile(t, withoutJust, ".dockerignore")
	if strings.Contains(dockerignore, "justfile") {
		t.Fatalf(".dockerignore should not contain justfile:\n%s", dockerignore)
	}
}

func TestRenderedJustfileListsRecipesAndHidesDefault(t *testing.T) {
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir, Just: true})

	cmd := exec.Command("just", "--justfile", filepath.Join(dir, "justfile"), "--working-directory", dir, "--list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("just --list failed: %v\n%s", err, output)
	}
	got := string(output)
	for _, want := range []string{"serve *args", "migrate *args", "snapshot *args", "upgrade version"} {
		if !strings.Contains(got, want) {
			t.Fatalf("just --list missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "default") {
		t.Fatalf("just --list should hide default recipe:\n%s", got)
	}
}

func TestRenderedJustfileSyntaxSupportsDryRun(t *testing.T) {
	if _, err := exec.LookPath("just"); err != nil {
		t.Skip("just is not installed")
	}
	dir := renderFixture(t, Config{MigrationDir: defaultMigrationDir, Just: true})

	for _, args := range [][]string{
		{"--dry-run", "serve", "--", "--http", "127.0.0.1:8090"},
		{"--dry-run", "migrate", "--", "--dir", "custom migrations"},
		{"--dry-run", "snapshot", "--", "-y", "--", "--flag", "value with spaces"},
		{"--dry-run", "upgrade", "v0.39.5"},
	} {
		fullArgs := append([]string{"--justfile", filepath.Join(dir, "justfile"), "--working-directory", dir}, args...)
		cmd := exec.Command("just", fullArgs...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("just dry-run %#v failed: %v\n%s", args, err, output)
		}
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
