package initcli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestIsGoModuleRequiresModuleFirstLineAndGoVersion(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "go.mod", "module example.com/app\n\nrequire example.com/other v1.0.0\n")
	ok, err := IsGoModule(dir)
	if err != nil {
		t.Fatalf("IsGoModule returned error: %v", err)
	}
	if ok {
		t.Fatal("expected go.mod without go version to be rejected")
	}

	writeFile(t, dir, "go.mod", "require example.com/other v1.0.0\nmodule example.com/app\n\ngo 1.20\n")
	ok, err = IsGoModule(dir)
	if err != nil {
		t.Fatalf("IsGoModule returned error: %v", err)
	}
	if ok {
		t.Fatal("expected go.mod without module first line to be rejected")
	}

	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.20\n")
	ok, err = IsGoModule(dir)
	if err != nil {
		t.Fatalf("IsGoModule returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected valid go.mod to be accepted")
	}
}

func TestIsGoModuleAcceptsGoPatchVersion(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n\ngo 1.20.3\n")

	ok, err := IsGoModule(dir)
	if err != nil {
		t.Fatalf("IsGoModule returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected go patch version to be accepted")
	}
}

func TestCurrentModuleWithoutGoSumOrGoFilesUsesCurrentDirectory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/current\n\ngo 1.20\n")
	runner := &fakeRunner{}

	code, stdout, stderr := runForProject([]string{}, dir, runner)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr)
	}
	if stdout != "" || stderr != "" {
		t.Fatalf("expected no output, stdout=%q stderr=%q", stdout, stderr)
	}
	assertCalls(t, runner.calls, []commandCall{
		{dir: dir, name: "go", args: []string{"get", "github.com/pocketbase/pocketbase@latest"}},
	})
}

func TestCurrentModuleWithGoSumRequiresForce(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/current\n\ngo 1.20\n")
	writeFile(t, dir, "go.sum", "")
	runner := &fakeRunner{}

	code, _, stderr := runForProject([]string{}, dir, runner)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "This directory is already initialized as a Go module.") {
		t.Fatalf("stderr missing force error: %q", stderr)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no commands, got %#v", runner.calls)
	}
}

func TestCurrentModuleWithGoFilesRequiresForce(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/current\n\ngo 1.20\n")
	writeFile(t, dir, "main.go", "package main\n")
	runner := &fakeRunner{}

	code, _, stderr := runForProject([]string{}, dir, runner)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "This directory is already initialized as a Go module.") {
		t.Fatalf("stderr missing force error: %q", stderr)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("expected no commands, got %#v", runner.calls)
	}
}

func TestCurrentModuleWithForcePrintsWarning(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/current\n\ngo 1.20\n")
	writeFile(t, dir, "go.sum", "")
	runner := &fakeRunner{}

	code, stdout, stderr := runForProject([]string{"--force"}, dir, runner)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Since the `--force` flag was provided") {
		t.Fatalf("stdout missing force warning: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	assertCalls(t, runner.calls, []commandCall{
		{dir: dir, name: "go", args: []string{"get", "github.com/pocketbase/pocketbase@latest"}},
	})
}

func TestMissingModuleNameOutsideGoModuleUsesBuildModulePath(t *testing.T) {
	dir := t.TempDir()
	runner := &fakeRunner{}

	code, _, stderr := runForProject([]string{}, dir, runner)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	for _, want := range []string{
		"This directory is not initialized as a Go module.",
		"`go run example.test/pb-init myproject`",
		"`go run example.test/pb-init github.com/username/myproject`",
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q: %q", want, stderr)
		}
	}
}

func TestModuleNameCreatesLastPathDirectoryAndRunsGoModInit(t *testing.T) {
	dir := t.TempDir()
	runner := &fakeRunner{}
	runner.onRun = func(call commandCall) (string, error) {
		if call.name == "go" && reflect.DeepEqual(call.args, []string{"mod", "init", "github.com/crmin/test-data"}) {
			writeFile(t, call.dir, "go.mod", "module github.com/crmin/test-data\n\ngo 1.20\n")
		}
		return "", nil
	}

	code, _, stderr := runForProject([]string{"github.com/crmin/test-data"}, dir, runner)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr)
	}
	target := filepath.Join(dir, "test-data")
	assertCalls(t, runner.calls, []commandCall{
		{dir: target, name: "go", args: []string{"mod", "init", "github.com/crmin/test-data"}},
		{dir: target, name: "go", args: []string{"get", "github.com/pocketbase/pocketbase@latest"}},
	})
}

func TestGoModInitFailureWritesCommandOutputToStderr(t *testing.T) {
	dir := t.TempDir()
	runner := &fakeRunner{output: "go mod init failed\n", err: errors.New("exit status 1")}

	code, stdout, stderr := runForProject([]string{"example.com/app"}, dir, runner)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if stderr != "go mod init failed\n" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestGoGetFailureWritesCommandOutputToStderr(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/current\n\ngo 1.20\n")
	runner := &fakeRunner{output: "go get failed\n", err: errors.New("exit status 1")}

	code, stdout, stderr := runForProject([]string{}, dir, runner)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout, got %q", stdout)
	}
	if stderr != "go get failed\n" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestJSVMRunsPluginGoGetForTransitiveDependencies(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/current\n\ngo 1.20\n")
	runner := &fakeRunner{}

	code, _, stderr := runForProject([]string{"--jsvm", "--pb-version=v0.39.5"}, dir, runner)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%q", code, stderr)
	}
	assertCalls(t, runner.calls, []commandCall{
		{dir: dir, name: "go", args: []string{"get", "github.com/pocketbase/pocketbase@v0.39.5"}},
		{dir: dir, name: "go", args: []string{"get", "github.com/pocketbase/pocketbase/plugins/jsvm@v0.39.5"}},
	})
}

func runForProject(args []string, dir string, runner *fakeRunner) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(args, Env{
		Stdout:      &stdout,
		Stderr:      &stderr,
		WorkDir:     dir,
		Templates:   testTemplateFS(),
		Runner:      runner,
		CommandPath: "example.test/pb-init",
	})
	return code, stdout.String(), stderr.String()
}

type commandCall struct {
	dir  string
	name string
	args []string
}

type fakeRunner struct {
	calls  []commandCall
	output string
	err    error
	onRun  func(commandCall) (string, error)
}

func (r *fakeRunner) Run(dir string, name string, args ...string) (string, error) {
	call := commandCall{dir: dir, name: name, args: append([]string(nil), args...)}
	r.calls = append(r.calls, call)
	if r.onRun != nil {
		return r.onRun(call)
	}
	return r.output, r.err
}

func assertCalls(t *testing.T, got []commandCall, want []commandCall) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected calls:\nwant: %#v\n got: %#v", want, got)
	}
}

func writeFile(t *testing.T, dir string, name string, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
