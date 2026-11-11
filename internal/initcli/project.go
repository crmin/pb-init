package initcli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	pocketBaseModule       = "ganigeorgiev.com/pocketbase/pocketbase"
	legacyPocketBaseModule = "github.com/pocketbase/pocketbase"

	forceRequiredMessage = "This directory is already initialized as a Go module. To initialize the current directory as a PocketBase project, run the command again with the `--force` flag.\n**Warning**: This may overwrite or damage your existing project.\n"
	forceProceedMessage  = "This directory is already initialized as a Go module. Since the `--force` flag was provided, PocketBase project initialization will proceed.\nWarning: Existing project files may be overwritten or corrupted.\n"
)

var goVersionLinePattern = regexp.MustCompile(`^go \d+\.\d+(?:\.\d+)?$`)

// Project describes the target PocketBase project module.
type Project struct {
	Dir              string
	RelativeDir      string
	ModulePath       string
	PocketBaseModule string
	FromModuleName   bool
}

// ProjectEnv contains dependencies for project preparation.
type ProjectEnv struct {
	WorkDir string
	Stdout  io.Writer
	Command string
	Runner  CommandRunner
}

// InitError is a user-facing initialization error.
type InitError struct {
	Message string
}

func (e *InitError) Error() string {
	return e.Message
}

// CommandError preserves external command output for stderr forwarding.
type CommandError struct {
	Output string
	Err    error
}

func (e *CommandError) Error() string {
	if e.Output != "" {
		return e.Output
	}
	return e.Err.Error()
}

// PrepareProject resolves or creates the target Go module and installs PocketBase.
func PrepareProject(cfg Config, env ProjectEnv) (Project, error) {
	project, err := resolveProject(cfg, env)
	if err != nil {
		return Project{}, err
	}

	modulePath, version, err := resolvePocketBase(project.Dir, cfg.PBVersion, env.Runner)
	if err != nil {
		return Project{}, err
	}
	project.PocketBaseModule = modulePath

	logStep(env.Stdout, "Installing PocketBase SDK: %s@%s", modulePath, version)
	if err := runGoGetPackage(project.Dir, modulePath, version, env.Runner); err != nil {
		return Project{}, err
	}
	if cfg.JSVM {
		jsvmPackage := modulePath + "/plugins/jsvm"
		logStep(env.Stdout, "Installing PocketBase JSVM plugin: %s@%s", jsvmPackage, version)
		if err := runGoGetPackage(project.Dir, jsvmPackage, version, env.Runner); err != nil {
			return Project{}, err
		}
	}

	project.ModulePath, err = ReadModulePath(project.Dir)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

type moduleVersion struct {
	Version string
	GoMod   string
}

func resolvePocketBase(dir string, query string, runner CommandRunner) (string, string, error) {
	queryPath := pocketBaseModule
	tryLegacy := false
	switch query {
	case defaultPBVersion:
	case "patch", "upgrade":
		path, err := currentPocketBaseModule(dir, runner)
		if err != nil {
			return "", "", err
		}
		queryPath = path
	default:
		tryLegacy = true
	}

	var release moduleVersion
	err := readGoJSON(dir, runner, &release, "list", "-m", "-json", queryPath+"@"+query)
	if tryLegacy && versionNotFound(err) {
		if legacyErr := readGoJSON(dir, runner, &release, "list", "-m", "-json", legacyPocketBaseModule+"@"+query); legacyErr != nil {
			return "", "", errors.Join(err, legacyErr)
		}
		err = nil
	}
	if err != nil {
		return "", "", err
	}
	if release.Version == "" || release.GoMod == "" {
		return "", "", fmt.Errorf("missing PocketBase version or go.mod for %q", query)
	}

	// go list reports the queried path, even when go.mod declares the older path.
	modulePath, err := readModulePath(release.GoMod)
	if err != nil {
		return "", "", fmt.Errorf("read PocketBase module path for %s: %w", release.Version, err)
	}
	if modulePath != pocketBaseModule && (modulePath != legacyPocketBaseModule || query == defaultPBVersion) {
		return "", "", fmt.Errorf("unexpected PocketBase module path %q for %s", modulePath, release.Version)
	}
	return modulePath, release.Version, nil
}

func currentPocketBaseModule(dir string, runner CommandRunner) (string, error) {
	var mod struct {
		Require []struct{ Path string }
	}
	if err := readGoJSON(dir, runner, &mod, "mod", "edit", "-json"); err != nil {
		return "", err
	}
	modulePath := pocketBaseModule
	for _, req := range mod.Require {
		switch req.Path {
		case pocketBaseModule:
			return pocketBaseModule, nil
		case legacyPocketBaseModule:
			modulePath = legacyPocketBaseModule
		}
	}
	return modulePath, nil
}

func versionNotFound(err error) bool {
	var commandErr *CommandError
	return errors.As(err, &commandErr) &&
		(strings.Contains(commandErr.Output, "unknown revision") ||
			strings.Contains(commandErr.Output, "no matching versions"))
}

func readGoJSON(dir string, runner CommandRunner, dst any, args ...string) error {
	output, err := runner.Output(dir, "go", args...)
	if err != nil {
		return newCommandError(output, err)
	}
	if err := json.Unmarshal([]byte(output), dst); err != nil {
		return fmt.Errorf("decode go %s output: %w", strings.Join(args, " "), err)
	}
	return nil
}

// IsGoModule checks the repository-local Go module contract from SPEC.md.
func IsGoModule(dir string) (bool, error) {
	content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "module") {
		return false, nil
	}

	for _, line := range lines {
		if goVersionLinePattern.MatchString(strings.TrimSpace(line)) {
			return true, nil
		}
	}

	return false, nil
}

// ReadModulePath reads the module path from go.mod.
func ReadModulePath(dir string) (string, error) {
	return readModulePath(filepath.Join(dir, "go.mod"))
}

func readModulePath(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("module path not found in go.mod")
}

func resolveProject(cfg Config, env ProjectEnv) (Project, error) {
	workDir, err := filepath.Abs(env.WorkDir)
	if err != nil {
		return Project{}, err
	}

	if cfg.ModuleName != "" {
		return createModuleProject(cfg, workDir, env)
	}

	isModule, err := IsGoModule(workDir)
	if err != nil {
		return Project{}, err
	}
	if !isModule {
		return Project{}, &InitError{Message: missingModuleNameMessage(env.Command)}
	}

	logStep(env.Stdout, "Using Go module directory: %s", workDir)

	requiresForce, err := currentModuleRequiresForce(workDir)
	if err != nil {
		return Project{}, err
	}
	if requiresForce {
		if !cfg.Force {
			return Project{}, &InitError{Message: forceRequiredMessage}
		}
		fmt.Fprint(env.Stdout, forceProceedMessage)
	}

	return Project{Dir: workDir, RelativeDir: "."}, nil
}

func createModuleProject(cfg Config, workDir string, env ProjectEnv) (Project, error) {
	dirName := path.Base(cfg.ModuleName)
	targetDir := filepath.Join(workDir, dirName)
	relativeDir, err := filepath.Rel(workDir, targetDir)
	if err != nil {
		relativeDir = targetDir
	}

	isModule, err := IsGoModule(targetDir)
	if err != nil {
		return Project{}, err
	}
	if isModule {
		logStep(env.Stdout, "Using Go module directory: %s", targetDir)

		requiresForce, err := currentModuleRequiresForce(targetDir)
		if err != nil {
			return Project{}, err
		}
		if requiresForce {
			if !cfg.Force {
				return Project{}, &InitError{Message: forceRequiredMessage}
			}
			fmt.Fprint(env.Stdout, forceProceedMessage)
		}

		return Project{Dir: targetDir, RelativeDir: relativeDir, FromModuleName: true}, nil
	}

	logStep(env.Stdout, "Preparing module directory: %s", targetDir)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Project{}, err
	}

	logStep(env.Stdout, "Initializing Go module: %s", cfg.ModuleName)
	if err := runCommand(targetDir, env.Runner, "go", "mod", "init", cfg.ModuleName); err != nil {
		return Project{}, err
	}

	return Project{Dir: targetDir, RelativeDir: relativeDir, FromModuleName: true}, nil
}

func currentModuleRequiresForce(dir string) (bool, error) {
	if _, err := os.Stat(filepath.Join(dir, "go.sum")); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	goFiles, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return false, err
	}

	return len(goFiles) > 0, nil
}

func runGoGetPackage(dir string, pkg string, version string, runner CommandRunner) error {
	return runCommand(dir, runner, "go", "get", pkg+"@"+version)
}

func runGoModTidy(dir string, runner CommandRunner) error {
	return runCommand(dir, runner, "go", "mod", "tidy")
}

func runCommand(dir string, runner CommandRunner, name string, args ...string) error {
	output, err := runner.Run(dir, name, args...)
	if err == nil {
		return nil
	}
	return newCommandError(output, err)
}

func newCommandError(output string, err error) *CommandError {
	if output == "" {
		output = err.Error() + "\n"
	}
	return &CommandError{Output: output, Err: err}
}

func missingModuleNameMessage(command string) string {
	return "This directory is not initialized as a Go module. To initialize the current directory as a PocketBase project, please provide a module name as an argument.\n\nExample:\n- `go run " + command + " myproject`\n- `go run " + command + " github.com/username/myproject`\n"
}

func logStep(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, format+"\n", args...)
}
