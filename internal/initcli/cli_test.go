package initcli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseHelpIgnoresModuleName(t *testing.T) {
	cfg, err := ParseArgs([]string{"example.com/app", "--help"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !cfg.Help {
		t.Fatal("expected help flag")
	}
	if cfg.ModuleName != "" {
		t.Fatalf("expected module name to be ignored, got %q", cfg.ModuleName)
	}
}

func TestParseShortFlagBundleExpandsDockerAutoMigrationJSVM(t *testing.T) {
	cfg, err := ParseArgs([]string{"-dmj"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !cfg.Docker || !cfg.AutoMigration || !cfg.JSVM {
		t.Fatalf("expected -dmj to enable docker, auto migration, and jsvm: %+v", cfg)
	}
}

func TestParseShortBundleRejectsHelp(t *testing.T) {
	err := parseErr(t, []string{"-dh"})
	assertErrMessage(t, err, "Invalid flag: -h\nCannot use -h in a short flag bundle.")
}

func TestParseShortBundleRejectsRecommend(t *testing.T) {
	err := parseErr(t, []string{"-dr"})
	assertErrMessage(t, err, "Invalid flag: -r\nCannot use -r in a short flag bundle.")
}

func TestParseShortBundleRejectsUnknownFlag(t *testing.T) {
	err := parseErr(t, []string{"-dmx"})
	assertErrMessage(t, err, "Invalid flag: -x")
}

func TestParseRecommendExpandsDockerAutoMigrationAndJust(t *testing.T) {
	cfg, err := ParseArgs([]string{"--recommend"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !cfg.Docker || !cfg.AutoMigration || !cfg.Just {
		t.Fatalf("expected --recommend to enable docker, auto migration, and justfile generation: %+v", cfg)
	}
}

func TestParseRecommendShortFlagExpandsDockerAutoMigrationAndJust(t *testing.T) {
	cfg, err := ParseArgs([]string{"-r"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !cfg.Docker || !cfg.AutoMigration || !cfg.Just {
		t.Fatalf("expected -r to enable docker, auto migration, and justfile generation: %+v", cfg)
	}
}

func TestParseJustFlag(t *testing.T) {
	cfg, err := ParseArgs([]string{"--just"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if !cfg.Just {
		t.Fatal("expected --just to enable justfile generation")
	}
}

func TestParseOptionsAcceptEqualsAndSeparateValues(t *testing.T) {
	cfg, err := ParseArgs([]string{"--migration-dir=internal/migrations", "--pb-version", "v0.39.5"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}
	if cfg.MigrationDir != "internal/migrations" {
		t.Fatalf("unexpected migration dir: %q", cfg.MigrationDir)
	}
	if cfg.PBVersion != "v0.39.5" {
		t.Fatalf("unexpected pb version: %q", cfg.PBVersion)
	}
}

func TestParseUnknownLongFlag(t *testing.T) {
	err := parseErr(t, []string{"--unknown"})
	assertErrMessage(t, err, "Invalid flag: --unknown")
}

func TestParseRejectsUnexpectedArgument(t *testing.T) {
	err := parseErr(t, []string{"example.com/app", "extra"})
	assertErrMessage(t, err, "Unexpected argument: extra")
}

func TestParseRejectsNonePBVersionToStderr(t *testing.T) {
	err := parseErr(t, []string{"--pb-version=none"})
	assertErrMessage(t, err, "Invalid --pb-version: none is not allowed. Provide a PocketBase version or omit --pb-version to use latest.")
}

func TestParseRejectsAbsoluteMigrationDirToStderr(t *testing.T) {
	err := parseErr(t, []string{"--migration-dir", filepath.Join(string(filepath.Separator), "tmp", "migrations")})
	assertErrMessage(t, err, "Invalid --migration-dir: absolute paths are not allowed. Use a child path relative to the PocketBase project module directory.")
}

func TestParseRejectsCurrentMigrationDirToStderr(t *testing.T) {
	for _, arg := range []string{"--migration-dir=.", "--migration-dir=internal/./migrations"} {
		err := parseErr(t, []string{arg})
		assertErrMessage(t, err, "Invalid --migration-dir: current directory references (`.`) are not allowed. Use a child path relative to the PocketBase project module directory.")
	}
}

func TestParseRejectsParentMigrationDirToStderr(t *testing.T) {
	err := parseErr(t, []string{"--migration-dir=../migrations"})
	assertErrMessage(t, err, "Invalid --migration-dir: parent directory references (`..`) are not allowed. Use a child path relative to the PocketBase project module directory.")
}

func TestParseInvalidShortFlagUsesInputFlagCharacter(t *testing.T) {
	err := parseErr(t, []string{"-dmz"})
	assertErrMessage(t, err, "Invalid flag: -z")
}

func TestHelpMessageUsesLongRecommendExpansion(t *testing.T) {
	help := HelpMessage("github.com/crmin/pb-init")
	if strings.Contains(help, "Equivalent to -dm.") {
		t.Fatal("help should not describe recommend with a short flag bundle")
	}
	if strings.Contains(help, "Equivalent to --docker --auto-migration.") {
		t.Fatal("help should not describe recommend without --just")
	}
	if !strings.Contains(help, "Equivalent to --docker --auto-migration --just.") {
		t.Fatal("help should describe recommend with long flags")
	}
}

func TestHelpMessageClarifiesMigrationDirWithJSVM(t *testing.T) {
	help := HelpMessage("github.com/crmin/pb-init")
	for _, want := range []string{"compiled into the Go build", "JavaScript migrations remain in `pb_migrations`", "JavaScript hooks remain in `pb_hooks`"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q", want)
		}
	}
}

func TestHelpMessageDocumentsJustFlag(t *testing.T) {
	help := HelpMessage("github.com/crmin/pb-init")
	if !strings.Contains(help, "--just") {
		t.Fatalf("help missing --just flag:\n%s", help)
	}
	if !strings.Contains(help, "Generate a justfile with common PocketBase project commands.") {
		t.Fatalf("help missing just flag description:\n%s", help)
	}
}

func TestErrorsWriteToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--pb-version=none"}, Env{
		Stdout:      &stdout,
		Stderr:      &stderr,
		CommandPath: "github.com/crmin/pb-init",
	})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "Invalid --pb-version: none is not allowed.") {
		t.Fatalf("stderr missing parse error: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "PocketBase project initializer") {
		t.Fatalf("stderr missing help message: %q", stderr.String())
	}
}

func parseErr(t *testing.T, args []string) *UsageError {
	t.Helper()
	_, err := ParseArgs(args)
	if err == nil {
		t.Fatalf("expected parse error for args %#v", args)
	}
	usageErr, ok := err.(*UsageError)
	if !ok {
		t.Fatalf("expected UsageError, got %T: %v", err, err)
	}
	return usageErr
}

func assertErrMessage(t *testing.T, err *UsageError, want string) {
	t.Helper()
	if err.Message != want {
		t.Fatalf("unexpected message:\nwant: %q\n got: %q", want, err.Message)
	}
}
