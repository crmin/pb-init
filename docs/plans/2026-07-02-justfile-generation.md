# justfile 생성 옵션 추가 계획

## 후속 변경

- 2026-07-02 후속 요구에 따라 `--recommend`은 `--docker --auto-migration --just`와 동일하게 변경됐다.
- 2026-07-02 후속 요구에 따라 generated `justfile`의 `snapshot` 정리 대상은 `migrations` 고정값이 아니라 `{{.MigrationDir}}`로 렌더링되는 configured migration directory다.

## 목표와 현재 동작

- 목표: `pb-init` 실행 옵션에 `--just` flag를 추가해 생성 대상 PocketBase module root에 `justfile`을 작성한다.
- 목표: `--docker --just`가 함께 전달된 경우에만 `.dockerignore`에 `justfile`을 추가한다. `--just`가 없으면 `.dockerignore`에 `justfile`을 추가하지 않는다.
- 목표: 생성된 `justfile`은 `serve`, `migrate`, `snapshot`, `upgrade` 명령을 제공하고, `default`는 `[private]`로 숨긴 채 `just --list`를 실행한다.
- 목표: 요구된 justfile 내용은 계획 단계에서 고정하고, subagent로 just 문법과 bash script 유효성을 검증한다.
- 현재 동작:
  - `Config`에는 `--just`에 대응하는 설정값이 없다.
  - `parseLongFlag`는 `--docker`, `--auto-migration`, `--jsvm`, `--cgo-enabled`, `--recommend`, `--migration-dir`, `--pb-version`만 처리한다.
  - `RenderProject`는 `main.go`, migration `init.go`, 선택적 Docker 파일, `.gitignore`만 생성한다.
  - `templates/`에는 `justfile` 템플릿이 없고 embed 검증 대상에도 포함되지 않는다.
  - `.dockerignore.tmpl`은 binary name만 조건부 렌더링하며 `justfile` 조건이 없다.

## 관련 파일과 코드 위치

- `internal/initcli/cli.go`
  - `Config`: `Just bool` 추가 위치.
  - `ParseArgs`, `parseLongFlag`: `--just` long flag 파싱 추가 위치.
  - `HelpMessage`: `--just` 설명 추가 위치.
- `internal/initcli/render.go`
  - template constant: `templates/justfile.tmpl` 추가 위치.
  - `RenderProject`: `cfg.Just`일 때 module root에 `justfile` 렌더링 추가 위치.
  - `renderData`: `.dockerignore.tmpl` 조건 분기를 위한 `Justfile` bool 추가 위치.
  - `templateData`: `cfg.Just` 전달 위치.
- `templates/justfile.tmpl`
  - 신규 template. 이번 계획에서 고정한 justfile 본문을 담는다.
- `templates/.dockerignore.tmpl`
  - `{{ if .Justfile }}justfile{{ end }}` 조건 추가 위치.
- `main_test.go`
  - embedded template 목록에 `templates/justfile.tmpl` 추가.
- `internal/initcli/cli_test.go`
  - `--just` 파싱과 help 문구 테스트 추가.
- `internal/initcli/render_test.go`
  - justfile 생성/미생성, `.dockerignore` 조건 렌더링, optional just 문법 smoke 테스트 추가.
- `SPEC.md`
  - Optional/Flags와 동작 section에 `--just`, 생성 파일, `.dockerignore` 조건, just command 계약 반영.
- `README.md`
  - Options/Flags, Generated Files, Examples 또는 Quick Start에 `--just` 사용법 반영.
- `docs/works/justfile-generation/`
  - 계획, 검증, 결정, 이슈 기록.

## 현재 계약 또는 명세 근거

- `SPEC.md`는 `--docker`가 전달된 경우에만 `Dockerfile`, `.dockerignore`를 생성한다고 정의한다.
- `SPEC.md`는 `.dockerignore.tmpl`의 `BinaryName` 조건 렌더링만 정의하고 있으며 `justfile` 조건은 아직 없다.
- `SPEC.md`는 templates directory의 모든 파일이 embed되어 build 후 binary에 포함되어야 한다고 정의한다.
- 사용자 요청은 신규 계약으로 `--just` flag, module root `justfile` 생성, `.dockerignore`의 조건부 `justfile` 항목, just 명령 동작, bash shebang, command echo 숨김, subagent 검증을 요구한다.
- `AGENTS.md`는 모든 task에서 `working-docs` skill로 작업 로그와 결정을 기록하라고 요구한다.
- `AGENTS.md`는 `RTK.md`를 참조하지만 현재 작업트리에서 `RTK.md`는 발견되지 않았다.

## 제안 변경 사항

### `internal/initcli/cli.go`

- `Config`에 `Just bool` 필드를 추가한다.
- `parseLongFlag`에 `--just` case를 추가해 `cfg.Just = true`로 설정한다.
- `HelpMessage`의 Flags section에 `--just` 설명을 추가한다.
  - 예상 문구: `--just Generate a justfile with common PocketBase project commands.`
- short flag는 추가하지 않는다. 사용자 요청에 short flag가 없고, 기존 short bundle 계약을 넓히지 않는 것이 범위가 작다.
- `--recommend`은 후속 요구에 따라 `--docker --auto-migration --just`를 의미한다.

### `internal/initcli/render.go`

- `templateJustfile = "templates/justfile.tmpl"` 상수를 추가한다.
- `RenderProject`에서 `cfg.Just`가 true일 때 `filepath.Join(project.Dir, "justfile")`에 `templates/justfile.tmpl`을 렌더링한다.
- `renderData`에 `Justfile bool`을 추가하고 `templateData`에서 `cfg.Just`를 전달한다.
- `.dockerignore`는 기존대로 `cfg.Docker`일 때만 생성한다.
  - `cfg.Docker && cfg.Just`: `.dockerignore`에 `justfile` 포함.
  - `cfg.Docker && !cfg.Just`: `.dockerignore`에 `justfile` 미포함.
  - `!cfg.Docker && cfg.Just`: `justfile`만 생성하고 `.dockerignore`는 생성하지 않음.

### `templates/justfile.tmpl`

계획 단계에서 다음 내용으로 고정한다.

```just
#!/usr/bin/env bash
set positional-arguments := true

# List available recipes and short descriptions.
[private]
default:
    @just --list

# Start the PocketBase server and forward arguments to `go run . serve`.
serve *args:
    #!/usr/bin/env bash
    set -euo pipefail
    go run . serve "$@"

# Create a collection migration snapshot with PocketBase.
migrate *args:
    #!/usr/bin/env bash
    set -euo pipefail
    go run . migrate collections "$@"

# Create a collection snapshot and keep only the newest migration file.
snapshot *args:
    #!/usr/bin/env bash
    set -euo pipefail

    yes=false
    forward_rest=false
    migrate_args=()
    delete_files=()
    migration_dir="{{.MigrationDir}}"

    for arg in "$@"; do
        if [[ "$forward_rest" == true ]]; then
            migrate_args+=("$arg")
            continue
        fi

        case "$arg" in
            -y)
                yes=true
                ;;
            --)
                forward_rest=true
                ;;
            *)
                migrate_args+=("$arg")
                ;;
        esac
    done

    if [[ ${#migrate_args[@]} -eq 0 ]]; then
        printf 'y\n' | go run . migrate collections
    else
        printf 'y\n' | go run . migrate collections "${migrate_args[@]}"
    fi

    if [[ ! -d "$migration_dir" ]]; then
        exit 0
    fi

    latest_file=""
    latest_ts=""

    while IFS= read -r file; do
        name="${file##*/}"
        ts="${name%%_*}"

        if [[ "$name" == *_*.go && "$ts" =~ ^[0-9]+$ ]]; then
            if [[ -z "$latest_file" || "$ts" -ge "$latest_ts" ]]; then
                latest_file="$file"
                latest_ts="$ts"
            fi
        fi
    done < <(find "$migration_dir" -maxdepth 1 -type f -name '*.go' -print | sort)

    while IFS= read -r file; do
        if [[ -z "$latest_file" || "$file" != "$latest_file" ]]; then
            delete_files+=("$file")
        fi
    done < <(find "$migration_dir" -maxdepth 1 -type f -name '*.go' -print | sort)

    if [[ ${#delete_files[@]} -eq 0 ]]; then
        exit 0
    fi

    if [[ "$yes" != true ]]; then
        printf 'The following files will be deleted:\n'
        printf '    - %s\n' "${delete_files[@]}"

        while true; do
            read -r -p 'The following files will be deleted. Continue? (Y/n): ' answer
            normalized=$(printf '%s' "$answer" | tr '[:upper:]' '[:lower:]')

            case "$normalized" in
                y)
                    break
                    ;;
                n)
                    printf 'Collection snapshot creation cancelled by user.\n'
                    exit 0
                    ;;
                *)
                    ;;
            esac
        done
    fi

    rm -f -- "${delete_files[@]}"

# Upgrade the PocketBase Go module dependency.
upgrade version="":
    #!/usr/bin/env bash
    set -euo pipefail

    version="$1"

    case "$version" in
        "")
            go get -u github.com/pocketbase/pocketbase
            ;;
        latest)
            go get -u github.com/pocketbase/pocketbase@latest
            ;;
        none)
            printf 'Invalid version: "none" is not allowed.\n' >&2
            exit 1
            ;;
        v*)
            go get -u "github.com/pocketbase/pocketbase@$version"
            ;;
        [0-9]*)
            go get -u "github.com/pocketbase/pocketbase@v$version"
            ;;
        *)
            printf 'Unsupported version format. Supported values are:\n' >&2
            printf '    - latest\n' >&2
            printf '    - a.b.c (e.g. 0.39.5)\n' >&2
            printf '    - va.b.c (e.g. v0.39.5)\n' >&2
            exit 1
            ;;
    esac
```

고정된 해석:

- `set positional-arguments := true`를 사용해 `serve`, `migrate`, `snapshot`의 공백 포함 인자를 `"$@"`로 보존한다.
- shebang recipe는 실제 실행 시 just가 recipe 본문을 echo하지 않으므로 stdout/stderr는 실행 명령의 출력만 표시된다.
- `snapshot`에서 `--` 앞의 `-y`는 recipe 전용 확인 생략 flag로 처리한다. `--` 뒤의 값은 `-y`라도 PocketBase migrate 인자로 전달한다.
- `snapshot`은 configured migration directory의 `*.go` 중 numeric timestamp prefix가 가장 큰 파일 하나를 최신 migration으로 보고, 그 외 모든 `.go` 파일은 삭제 대상으로 본다. 따라서 `init.go`와 non-numeric `.go`도 삭제 대상이다.
- prompt 입력은 명시 요구대로 `y`/`Y`만 진행, `n`/`N`만 취소, 그 외 값과 빈 입력은 재프롬프트한다.
- `upgrade`는 사용자 요구의 prefix 규칙을 따른다. `latest`, `none`, `v`로 시작, 숫자로 시작, 그 외 값을 분기하며 semver 정규식 검증은 추가하지 않는다.

### `templates/.dockerignore.tmpl`

- 조건부 항목을 추가한다.

```gotemplate
{{ if .Justfile }}justfile
{{ end }}
```

- 위치는 `*.md` 또는 binary name 주변 중 하나로 두되, 결과에 빈 줄이 과도하게 늘어나지 않도록 템플릿 newline을 확인한다.

### `main_test.go`

- `TestEmbeddedTemplatesIncludeAllRequiredFiles`에 `templates/justfile.tmpl`을 추가한다.

### `internal/initcli/cli_test.go`

- `TestParseJustFlag`를 추가해 `ParseArgs([]string{"--just"})`가 `cfg.Just`를 true로 설정하는지 확인한다.
- `TestHelpMessageDocumentsJustFlag`를 추가해 help message에 `--just` 설명이 포함되는지 확인한다.
- 기존 short flag bundle 테스트는 유지하고 `--just`가 short bundle에 섞이지 않는 계약을 변경하지 않는다.

### `internal/initcli/render_test.go`

- `TestRenderWritesJustfileWhenJustEnabled`
  - `Config{MigrationDir: defaultMigrationDir, Just: true}` 렌더링 후 `justfile` 존재와 주요 recipe 문자열을 확인한다.
- `TestRenderSkipsJustfileWhenJustDisabled`
  - 기본 config 렌더링 후 `justfile`이 생성되지 않는지 확인한다.
- `TestRenderDockerignoreIncludesJustfileOnlyWhenJustEnabled`
  - `Docker: true, Just: true`일 때 `.dockerignore`에 `justfile` 포함.
  - `Docker: true, Just: false`일 때 `.dockerignore`에 `justfile` 미포함.
- `TestRenderedJustfileListsRecipesAndHidesDefault`
  - `just` binary가 있으면 임시 module에서 `just --justfile <dir>/justfile --working-directory <dir> --list`를 실행해 `serve`, `migrate`, `snapshot`, `upgrade`는 보이고 `default`는 보이지 않는지 확인한다.
  - `just` binary가 없으면 skip한다.
- `TestRenderedJustfileSyntaxSupportsDryRun`
  - `just` binary가 있으면 `serve`, `migrate`, `snapshot`, `upgrade`의 dry-run이 parse 가능한지 확인한다.
  - 실제 `go`, `./pocketbase`, `rm` 실행은 하지 않는다.

### `SPEC.md`

- Flags section에 `--just`를 추가한다.
- 동작 section에 `--just` 시 module root `justfile` 생성 계약을 추가한다.
- Docker file 생성 section의 `.dockerignore.tmpl` variables에 `{{.Justfile}}` 또는 동등한 조건을 추가하고, `--just`가 있을 때만 `justfile`이 렌더링된다고 명시한다.
- just command 동작을 사용자 요청의 문구와 동일한 수준으로 반영한다.

### `README.md`

- Flags 또는 Generated Files에 `--just`를 추가한다.
- 생성된 just commands를 간략히 설명한다.
- `--docker --just`일 때 `.dockerignore`에 `justfile`이 포함된다는 조건을 설명한다.

### branch, commit, merge

- 사용자 승인 후 현재 `main`에서 `codex/justfile-generation` 브랜치를 생성한다.
- 구현, 테스트, 문서 업데이트를 같은 원자 작업 단위로 commit한다.
- 검증 완료 후 `main`에 fast-forward merge 또는 rebase 방식으로 반영한다.
- 현재 `main`은 `origin/main`보다 2커밋 앞서 있으므로, 새 브랜치는 현재 로컬 `main` HEAD 기준으로 만든다.

## TDD 계획

1. `TestParseJustFlag`를 추가하고 실패를 확인한다.
2. `TestHelpMessageDocumentsJustFlag`를 추가하고 실패를 확인한다.
3. `TestRenderWritesJustfileWhenJustEnabled`를 추가하고 실패를 확인한다.
4. `TestRenderSkipsJustfileWhenJustDisabled`를 추가하고 실패를 확인한다.
5. `TestRenderDockerignoreIncludesJustfileOnlyWhenJustEnabled`를 추가하고 실패를 확인한다.
6. `TestEmbeddedTemplatesIncludeAllRequiredFiles`에 `templates/justfile.tmpl` 기대값을 추가하고 실패를 확인한다.
7. 가능하면 `TestRenderedJustfileListsRecipesAndHidesDefault`, `TestRenderedJustfileSyntaxSupportsDryRun`을 skip 가능한 smoke test로 추가한다.
8. 최소 구현으로 테스트를 통과시킨다.
9. `SPEC.md`, `README.md`, `docs/works/justfile-generation/`을 구현 내용과 일치하도록 업데이트한다.
10. 리팩터링 후 전체 테스트와 build를 다시 실행한다.

## 변경 후 기대 동작

- `go run github.com/crmin/pb-init myproject --just`는 `myproject/justfile`을 생성한다.
- `go run github.com/crmin/pb-init myproject --recommend`는 `--docker --auto-migration --just`를 적용한다.
- `--just`가 없으면 `justfile`은 생성되지 않는다.
- `--docker --just`는 `.dockerignore`에 `justfile`을 포함한다.
- `--docker`만 있고 `--just`가 없으면 `.dockerignore`에 `justfile`이 포함되지 않는다.
- `--just`만 있고 `--docker`가 없으면 `.dockerignore`는 기존 계약대로 생성되지 않는다.
- 생성된 module에서 `just`는 `just --list`를 실행하고 `default` recipe 자체는 목록에 표시하지 않는다.
- `just serve [args...]`는 `go run . serve [args...]`를 실행한다.
- `just migrate [args...]`는 `go run . migrate collections [args...]`를 실행한다.
- `just snapshot [-y] [-- args...]`는 collection snapshot을 만들고 configured migration directory의 최신 timestamp migration `.go` 하나만 남기며, `-y`가 없으면 삭제 전 확인 prompt를 표시한다.
- `just upgrade [version]`은 요구된 version 분기에 따라 `go get -u github.com/pocketbase/pocketbase...`를 실행하거나 지정 오류를 stderr로 출력한다.

## 예상 부작용과 호환성 위험

- `--just` 추가로 help output과 `Config` 구조가 확장된다.
- 생성된 justfile은 `just` binary가 설치된 환경에서만 사용할 수 있다. `pb-init` 자체 실행에는 `just`가 필요하지 않다.
- `snapshot`은 configured migration directory의 `*.go` 중 최신 numeric timestamp 파일 하나만 보존한다. 수동 Go migration 파일도 `.go`이면 삭제 대상이 되므로, 삭제 prompt와 `-y` 사용 여부가 중요하다.
- 삭제 대상 목록은 `    - {file}` bullet 형식으로 출력한다.
- prompt 문구는 요구된 문자열 `The following files will be deleted. Continue? (Y/n): `을 그대로 사용하지만, 빈 Enter는 진행으로 처리하지 않고 재프롬프트한다.
- shebang recipe는 just actual run에서 command body를 echo하지 않는다. `just --dry-run`은 검증 목적상 body를 출력하는 것이 정상이다.
- `upgrade`는 요구된 starts-with 분기를 그대로 따르므로 `vbad`, `1beta` 같은 값도 `go get`으로 전달된다. 이 경우 실패 여부는 Go command가 결정한다.

## 검증 단계

- 단위 테스트:

```sh
go test ./...
```

- 빌드:

```sh
go build ./...
```

- justfile syntax smoke:

```sh
just --justfile <generated-module>/justfile --working-directory <generated-module> --list
just --justfile <generated-module>/justfile --working-directory <generated-module> --dry-run serve -- --http "127.0.0.1:8090"
just --justfile <generated-module>/justfile --working-directory <generated-module> --dry-run migrate -- --dir "custom migrations"
just --justfile <generated-module>/justfile --working-directory <generated-module> --dry-run snapshot -- -y -- --flag "value with spaces"
just --justfile <generated-module>/justfile --working-directory <generated-module> --dry-run upgrade v0.39.5
```

- 생성 결과 smoke:

```sh
tmpdir=$(mktemp -d)
go run . example.com/just-smoke --just --docker --force
```

확인 항목:

- `just-smoke/justfile` 존재.
- `just-smoke/.dockerignore`에 `justfile` 존재.
- `just --list`에 `serve`, `migrate`, `snapshot`, `upgrade` 존재.
- `just --list`에 `default` 미표시.

## TODO 체크리스트

- [x] `planner`와 `working-docs` 지침 읽기.
- [x] 현재 CLI, 렌더링, 템플릿, 테스트 구조 조사.
- [x] `RTK.md` 존재 여부 확인.
- [x] justfile 후보 작성.
- [x] subagent 1차 검증 수행.
- [x] subagent 피드백 반영.
- [x] subagent 2차 검증 수행.
- [x] 계획 문서 작성.
- [x] 사용자 계획 승인 받기.
- [x] 별도 브랜치 `codex/justfile-generation` 생성.
- [x] 실패 테스트 추가 및 예상 실패 확인.
- [x] `--just` parser/help 구현.
- [x] justfile 렌더링과 `.dockerignore` 조건 구현.
- [x] `SPEC.md`, `README.md`, work docs 최신화.
- [x] 전체 테스트와 build 실행.
- [x] 필요 시 리팩터링 후 재검증.
- [x] 작업 브랜치 commit.
- [x] `main`에 merge 또는 rebase.
