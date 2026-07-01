package initcli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

const (
	defaultMigrationDir = "migrations"
	defaultPBVersion    = "latest"

	msgInvalidPBVersionNone = "Invalid --pb-version: none is not allowed. Provide a PocketBase version or omit --pb-version to use latest."
	msgInvalidMigrationAbs  = "Invalid --migration-dir: absolute paths are not allowed. Use a child path relative to the PocketBase project module directory."
	msgInvalidMigrationCur  = "Invalid --migration-dir: current directory references (`.`) are not allowed. Use a child path relative to the PocketBase project module directory."
	msgInvalidMigrationUp   = "Invalid --migration-dir: parent directory references (`..`) are not allowed. Use a child path relative to the PocketBase project module directory."
)

// Config is the parsed command-line contract for pb-init.
type Config struct {
	ModuleName    string
	MigrationDir  string
	PBVersion     string
	Help          bool
	Force         bool
	Docker        bool
	AutoMigration bool
	JSVM          bool
	CgoEnabled    bool
}

// Env contains process dependencies used by Run.
type Env struct {
	Stdout      io.Writer
	Stderr      io.Writer
	WorkDir     string
	Templates   fs.FS
	CommandPath string
}

// UsageError is an error that should be printed with the help message.
type UsageError struct {
	Message string
}

func (e *UsageError) Error() string {
	return e.Message
}

func (e *UsageError) Format(command string) string {
	return e.Message + "\n\n" + HelpMessage(command)
}

// Run executes the CLI and returns the intended process exit code.
func Run(args []string, env Env) int {
	stdout := env.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	stderr := env.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	command := env.CommandPath
	if command == "" {
		command = buildInfoCommandPath()
	}

	cfg, err := ParseArgs(args)
	if err != nil {
		if usageErr, ok := err.(*UsageError); ok {
			fmt.Fprint(stderr, usageErr.Format(command))
			return 1
		}

		fmt.Fprintln(stderr, err)
		return 1
	}

	if cfg.Help {
		fmt.Fprint(stdout, HelpMessage(command))
		return 0
	}

	return 0
}

// ParseArgs parses the pb-init command line.
func ParseArgs(args []string) (Config, error) {
	cfg := Config{
		MigrationDir: defaultMigrationDir,
		PBVersion:    defaultPBVersion,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			cfg.Help = true
			cfg.ModuleName = ""
			return cfg, nil
		case strings.HasPrefix(arg, "--"):
			var err error
			i, err = parseLongFlag(args, i, &cfg)
			if err != nil {
				return Config{}, err
			}
		case isShortFlag(arg):
			if err := parseShortFlag(arg, &cfg); err != nil {
				return Config{}, err
			}
		default:
			if cfg.ModuleName != "" {
				return Config{}, usageError("Unexpected argument: " + arg)
			}
			cfg.ModuleName = arg
		}
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// HelpMessage returns the static English usage text.
func HelpMessage(command string) string {
	return fmt.Sprintf(`PocketBase project initializer

Usage:
  go run %s [moduleName] [options] [flags]

Arguments:
  moduleName
    Optional Go module path. When provided, pb-init creates a directory named after the last path element and runs `+"`go mod init <moduleName>`"+` in it. When omitted, the current directory must already be a Go module.

Options:
  --migration-dir=<dirName>
    Directory for generated Go migration files that are compiled into the Go build. The value may be nested, for example `+"`internal/migrations`"+`, but it must be a child path relative to the PocketBase project module directory: absolute paths, current directory references (`+"`.`"+`), and parent directory references (`+"`..`"+`) are rejected. When `+"`--jsvm`"+` is used, this option still only controls Go migrations; JavaScript migrations remain in `+"`pb_migrations`"+` and JavaScript hooks remain in `+"`pb_hooks`"+`. Default: migrations
  --pb-version=<version>
    PocketBase SDK version passed to `+"`go get github.com/pocketbase/pocketbase@<version>`"+`. Default: latest

Flags:
  -h, --help
    Show this help message and exit.
  --force
    Initialize the current Go module even when go.sum or Go source files already exist.
  -d, --docker
    Generate Dockerfile and .dockerignore.
  -m, --auto-migration
    Enable PocketBase auto migration in generated code.
  -j, --jsvm
    Enable PocketBase JS hooks, JS migrations, and JS runtime support.
  --cgo-enabled
    Use CGO_ENABLED=1 in the generated Dockerfile when --docker is enabled.
  -r, --recommend
    Equivalent to --docker --auto-migration.
`, command)
}

func parseLongFlag(args []string, index int, cfg *Config) (int, error) {
	arg := args[index]
	name, value, hasValue := strings.Cut(arg, "=")

	switch name {
	case "--help":
		cfg.Help = true
		cfg.ModuleName = ""
	case "--force":
		cfg.Force = true
	case "--docker":
		cfg.Docker = true
	case "--auto-migration":
		cfg.AutoMigration = true
	case "--jsvm":
		cfg.JSVM = true
	case "--cgo-enabled":
		cfg.CgoEnabled = true
	case "--recommend":
		applyRecommend(cfg)
	case "--migration-dir":
		next, consumed, err := optionValue(args, index, value, hasValue, "--migration-dir")
		if err != nil {
			return index, err
		}
		cfg.MigrationDir = next
		if consumed {
			index++
		}
	case "--pb-version":
		next, consumed, err := optionValue(args, index, value, hasValue, "--pb-version")
		if err != nil {
			return index, err
		}
		cfg.PBVersion = next
		if consumed {
			index++
		}
	default:
		return index, usageError("Invalid flag: " + name)
	}

	return index, nil
}

func optionValue(args []string, index int, inlineValue string, hasInlineValue bool, name string) (string, bool, error) {
	if hasInlineValue {
		if inlineValue == "" {
			return "", false, usageError("Missing value for " + name + ".")
		}
		return inlineValue, false, nil
	}

	nextIndex := index + 1
	if nextIndex >= len(args) || strings.HasPrefix(args[nextIndex], "-") {
		return "", false, usageError("Missing value for " + name + ".")
	}

	return args[nextIndex], true, nil
}

func parseShortFlag(arg string, cfg *Config) error {
	if len(arg) == 2 {
		switch arg[1] {
		case 'h':
			cfg.Help = true
			cfg.ModuleName = ""
		case 'd':
			cfg.Docker = true
		case 'm':
			cfg.AutoMigration = true
		case 'j':
			cfg.JSVM = true
		case 'r':
			applyRecommend(cfg)
		default:
			return usageError("Invalid flag: -" + string(arg[1]))
		}
		return nil
	}

	for _, flag := range arg[1:] {
		switch flag {
		case 'h', 'r':
			return usageError(fmt.Sprintf("Invalid flag: -%c\nCannot use -%c in a short flag bundle.", flag, flag))
		case 'd':
			cfg.Docker = true
		case 'm':
			cfg.AutoMigration = true
		case 'j':
			cfg.JSVM = true
		default:
			return usageError("Invalid flag: -" + string(flag))
		}
	}

	return nil
}

func isShortFlag(arg string) bool {
	return strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && arg != "-"
}

func validateConfig(cfg Config) error {
	if cfg.PBVersion == "none" {
		return usageError(msgInvalidPBVersionNone)
	}

	if filepath.IsAbs(cfg.MigrationDir) {
		return usageError(msgInvalidMigrationAbs)
	}

	for _, part := range splitPathParts(cfg.MigrationDir) {
		switch part {
		case ".":
			return usageError(msgInvalidMigrationCur)
		case "..":
			return usageError(msgInvalidMigrationUp)
		}
	}

	return nil
}

func splitPathParts(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})
}

func applyRecommend(cfg *Config) {
	cfg.Docker = true
	cfg.AutoMigration = true
}

func usageError(message string) *UsageError {
	return &UsageError{Message: message}
}

func buildInfoCommandPath() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Path == "" {
		return "github.com/crmin/pb-init"
	}
	return info.Main.Path
}
