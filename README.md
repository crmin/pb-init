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

PocketBase SDK 버전 또는 Go 버전 쿼리를 지정합니다. 기본값은 `latest`입니다.

`latest`를 지정하면 `ganigeorgiev.com/pocketbase/pocketbase`에서 버전을 조회합니다. `v0.39.5`처럼 GitHub 경로를 선언한 버전은 `github.com/pocketbase/pocketbase`에서 설치합니다. 기존 프로젝트에서 `patch` 또는 `upgrade`를 지정하면 현재 PocketBase 의존성의 경로를 사용합니다. `none`은 사용할 수 없습니다.

`--force`로 기존 프로젝트를 새 경로로 전환하려면 기존 Go 코드와 마이그레이션의 PocketBase import도 바꿔야 합니다. GitHub 경로로 등록한 마이그레이션은 그 경로의 PocketBase 앱에서만 실행됩니다. 기존 import를 유지하려면 GitHub 경로를 선언한 버전을 지정합니다.

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

Equivalent to `--docker --auto-migration --just`.

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
go run . migrate collections [args...]
```

`just snapshot [-y] [-- args...]`

Create a collection snapshot with `go run . migrate collections` and keep only the newest Go migration file in the configured `--migration-dir` directory. Without `-y`, it prints the files that will be deleted as an indented bullet list and asks for confirmation.

`just upgrade [version]`

프로젝트를 생성할 때 선택한 모듈 경로에서 PocketBase를 업그레이드합니다. 버전을 생략하거나 `latest`, `0.41.0`, `v0.41.0`처럼 해당 경로의 버전을 지정할 수 있습니다.

기존 GitHub 경로에서 새 경로로 전환하려면 Go import와 `go.mod`의 PocketBase 의존성을 함께 바꿉니다. `justfile`을 사용한다면 `upgrade`가 참조하는 경로도 바꿉니다.

## Examples

Create a recommended project:

```sh
go run github.com/crmin/pb-init myproject --recommend
```

This enables Docker generation, auto migration, and justfile generation.

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
