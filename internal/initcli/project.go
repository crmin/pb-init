package initcli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	forceRequiredMessage = "This directory is already initialized as a Go module. To initialize the current directory as a PocketBase project, run the command again with the `--force` flag.\n**Warning**: This may overwrite or damage your existing project.\n"
	forceProceedMessage  = "This directory is already initialized as a Go module. Since the `--force` flag was provided, PocketBase project initialization will proceed.\nWarning: Existing project files may be overwritten or corrupted.\n"
)

var goVersionLinePattern = regexp.MustCompile(`^go \d+\.\d+(?:\.\d+)?$`)

// Project describes the target PocketBase project module.
type Project struct {
	Dir        string
	ModulePath string
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

	if err := runGoGet(project.Dir, cfg.PBVersion, env.Runner); err != nil {
		return Project{}, err
	}
	if cfg.JSVM {
		if err := runGoGetPackage(project.Dir, "github.com/pocketbase/pocketbase/plugins/jsvm", cfg.PBVersion, env.Runner); err != nil {
			return Project{}, err
		}
	}

	modulePath, err := ReadModulePath(project.Dir)
	if err != nil {
		return Project{}, err
	}
	project.ModulePath = modulePath

	return project, nil
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
	file, err := os.Open(filepath.Join(dir, "go.mod"))
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
	if cfg.ModuleName != "" {
		return createModuleProject(cfg.ModuleName, env.WorkDir, env.Runner)
	}

	isModule, err := IsGoModule(env.WorkDir)
	if err != nil {
		return Project{}, err
	}
	if !isModule {
		return Project{}, &InitError{Message: missingModuleNameMessage(env.Command)}
	}

	requiresForce, err := currentModuleRequiresForce(env.WorkDir)
	if err != nil {
		return Project{}, err
	}
	if requiresForce {
		if !cfg.Force {
			return Project{}, &InitError{Message: forceRequiredMessage}
		}
		fmt.Fprint(env.Stdout, forceProceedMessage)
	}

	return Project{Dir: env.WorkDir}, nil
}

func createModuleProject(moduleName string, workDir string, runner CommandRunner) (Project, error) {
	dirName := path.Base(moduleName)
	targetDir := filepath.Join(workDir, dirName)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return Project{}, err
	}

	if err := runCommand(targetDir, runner, "go", "mod", "init", moduleName); err != nil {
		return Project{}, err
	}

	return Project{Dir: targetDir}, nil
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

func runGoGet(dir string, version string, runner CommandRunner) error {
	return runGoGetPackage(dir, "github.com/pocketbase/pocketbase", version, runner)
}

func runGoGetPackage(dir string, pkg string, version string, runner CommandRunner) error {
	return runCommand(dir, runner, "go", "get", pkg+"@"+version)
}

func runCommand(dir string, runner CommandRunner, name string, args ...string) error {
	output, err := runner.Run(dir, name, args...)
	if err == nil {
		return nil
	}
	if output == "" {
		output = err.Error() + "\n"
	}
	return &CommandError{Output: output, Err: err}
}

func missingModuleNameMessage(command string) string {
	return "This directory is not initialized as a Go module. To initialize the current directory as a PocketBase project, please provide a module name as an argument.\n\nExample:\n- `go run " + command + " myproject`\n- `go run " + command + " github.com/username/myproject`\n"
}
