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

This creates `./myproject`, runs `go mod init myproject`, installs PocketBase, and writes the starter files.

Use a full module path when you want the generated module to keep that path:

```sh
go run github.com/crmin/pb-init github.com/username/myproject
```

This creates `./myproject`, but runs:

```sh
go mod init github.com/username/myproject
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

Allow initialization in a current Go module that already has `go.sum` or root-level Go files.

`--docker`, `-d`

Generate `Dockerfile` and `.dockerignore`.

`--auto-migration`, `-m`

Enable PocketBase auto migration in the generated Go code.

`--jsvm`, `-j`

Enable PocketBase JS hooks, JS migrations, and JS runtime support. This also creates empty `pb_migrations` and `pb_hooks` directories in the project module directory.

When used with `--docker`, the generated Dockerfile copies `pb_migrations` and `pb_hooks` into the final image.

`--cgo-enabled`

Use `CGO_ENABLED=1` in the generated Dockerfile. This only affects Dockerfile generation when `--docker` is used.

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

## Examples

Create a recommended project:

```sh
go run github.com/crmin/pb-init myproject --recommend
```

Create a Docker-ready project with JSVM and nested Go migrations:

```sh
go run github.com/crmin/pb-init github.com/username/myproject --docker --jsvm --migration-dir=internal/migrations
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
