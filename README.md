# pb-init

Initialize a Go-based PocketBase project from a remote module run.

```sh
go run github.com/crmin/pb-init [moduleName] [options] [flags]
```

## Quick Start

Create a new project directory and initialize it as a Go module:

```sh
go run github.com/crmin/pb-init myproject
```

This creates `./myproject`, runs `go mod init myproject`, installs PocketBase, writes the starter files, and runs `go mod tidy`.

Use a full module path when you want the generated module to keep that path:

```sh
go run github.com/crmin/pb-init github.com/username/myproject
```

This creates `./myproject`, but runs:

```sh
go mod init github.com/username/myproject
```

When initialization finishes, pb-init prints the module path and the next commands to run:

```text
PocketBase project initialized successfully: /absolute/path/to/myproject

Go to module directory:
    cd myproject

Start the server:
    go run . serve

Create a collection snapshot:
    go run . migrate collections

Create a superuser:
    go run . superuser create <user_email> <user_password>
```

## Initialize The Current Directory

When `moduleName` is omitted, the current directory must already be a Go module:

```sh
go mod init example.com/myproject
go run github.com/crmin/pb-init
```

If the current Go module already has `go.sum` or root-level `*.go` files, pb-init stops unless `--force` is provided:

```sh
go run github.com/crmin/pb-init --force
```

`--force` may overwrite or damage existing project files.

The same protection applies when `moduleName` points to a directory that is already a Go module.

## Options

`--migration-dir=<dirName>`

Directory for generated Go migration files. The value is relative to the generated PocketBase project module directory and must be a child path, such as `migrations` or `internal/migrations`.

Absolute paths, current directory references (`.`), and parent directory references (`..`) are rejected.

When `--jsvm` is used, `--migration-dir` still only controls Go migrations. JavaScript migrations use `pb_migrations`, and JavaScript hooks use `pb_hooks`.

Default: `migrations`

`--pb-version=<version>`

PocketBase SDK version passed to:

```sh
go get github.com/pocketbase/pocketbase@<version>
```

Default: `latest`

The value `none` is not allowed. Omit `--pb-version` to use `latest`.

## Flags

`--help`, `-h`

Print the help message and exit.

`--force`

Allow initialization in an existing Go module that already has `go.sum` or root-level Go files.

`--docker`, `-d`

Generate `Dockerfile` and `.dockerignore`.

`--auto-migration`, `-m`

Enable PocketBase auto migration in the generated Go code.

`--jsvm`, `-j`

Enable PocketBase JS hooks, JS migrations, and JS runtime support. This also creates empty `pb_migrations` and `pb_hooks` directories in the project module directory.

When used with `--docker`, the generated Dockerfile copies `pb_migrations` and `pb_hooks` into the final image.

`--cgo-enabled`

Use `CGO_ENABLED=1` in the generated Dockerfile. This only affects Dockerfile generation when `--docker` is used.

`--just`

Generate a `justfile` in the project module root with common PocketBase project commands.

`--recommend`, `-r`

Equivalent to `--docker --auto-migration`.

Short flags except `-h` and `-r` can be bundled:

```sh
go run github.com/crmin/pb-init myproject -dmj
```

This is equivalent to:

```sh
go run github.com/crmin/pb-init myproject --docker --auto-migration --jsvm
```

## Generated Files

pb-init writes these files and directories:

```text
main.go
.gitignore
<migration-dir>/init.go
```

With `--docker`:

```text
Dockerfile
.dockerignore
```

With `--jsvm`:

```text
pb_migrations/
pb_hooks/
```

With `--just`:

```text
justfile
```

When `--docker` and `--just` are used together, `.dockerignore` also includes `justfile`. Without `--just`, `.dockerignore` does not include it.

## Generated just Commands

`just`

List available recipes with short descriptions. The default recipe is private, so it does not appear in the list.

`just serve [args...]`

Run:

```sh
go run . serve [args...]
```

`just migrate [args...]`

Run:

```sh
./pocketbase migrate collections [args...]
```

`just snapshot [-y] [-- args...]`

Create a collection snapshot and keep only the newest Go migration file in `migrations/`. Without `-y`, it prints the files that will be deleted and asks for confirmation.

`just upgrade [version]`

Upgrade the PocketBase Go dependency. Use no version, `latest`, a version like `0.39.5`, or a `v`-prefixed version like `v0.39.5`.

## Examples

Create a recommended project:

```sh
go run github.com/crmin/pb-init myproject --recommend
```

Create a Docker-ready project with JSVM and nested Go migrations:

```sh
go run github.com/crmin/pb-init github.com/username/myproject --docker --jsvm --migration-dir=internal/migrations
```

Create a project with a generated justfile:

```sh
go run github.com/crmin/pb-init myproject --just
```

Use a specific PocketBase version:

```sh
go run github.com/crmin/pb-init myproject --pb-version=v0.39.5
```

Initialize an existing empty Go module:

```sh
go mod init example.com/current
go run github.com/crmin/pb-init --docker --auto-migration
```

## Error Output

Requested help is printed to stdout.

Errors are printed to stderr. When an external Go command fails, pb-init forwards the command output to stderr and exits with code 1.

Progress logs and the final next-step summary are printed to stdout. Commands and paths in the final summary are colorized.
