# justfile migration dir 및 recommend 확장 계획

## 목표와 현재 동작

- 목표: `just snapshot`의 정리 대상 디렉토리를 고정 `migrations`가 아니라 `--migration-dir`로 지정된 경로로 생성 시점에 templating한다.
- 목표: `templates/justfile.tmpl`은 현재 경로를 유지하고, 해당 템플릿 안에서 `{{.MigrationDir}}`를 사용해 generated `justfile`에 migration directory 값을 주입한다.
- 목표: `--recommend`와 `-r`이 기존 `--docker --auto-migration`에 더해 `--just`도 활성화하도록 변경한다.
- 현재 동작:
  - `templates/justfile.tmpl`의 `snapshot` recipe는 `migrations` 문자열을 직접 사용한다.
  - `--migration-dir=internal/migrations --just`로 생성된 프로젝트에서도 `just snapshot`은 `internal/migrations`가 아니라 `migrations`를 정리 대상으로 본다.
  - `applyRecommend`는 `cfg.Docker = true`, `cfg.AutoMigration = true`만 설정하고 `cfg.Just`는 설정하지 않는다.
  - `SPEC.md`, `README.md`, `HelpMessage`는 `--recommend`을 `--docker --auto-migration`과 동일하다고 설명한다.

## 관련 파일과 코드 위치

- `templates/justfile.tmpl`
  - `snapshot` recipe의 `migrations` hardcoding 제거.
  - `migration_dir="{{.MigrationDir}}"` 같은 bash 변수로 템플릿 값을 주입하고 `find "$migration_dir"` 형태로 사용.
- `internal/initcli/render.go`
  - `renderData.MigrationDir`는 이미 존재하므로 새 template data field는 필요 없을 것으로 예상.
  - `renderTemplateFile`은 `missingkey=error`를 사용하므로 `{{.MigrationDir}}` 사용은 기존 data로 검증된다.
- `internal/initcli/cli.go`
  - `applyRecommend`: `cfg.Just = true` 추가.
  - `HelpMessage`: `--recommend` 설명을 `--docker --auto-migration --just`로 변경.
- `internal/initcli/cli_test.go`
  - `TestParseRecommendExpandsDockerAutoMigration`를 `TestParseRecommendExpandsDockerAutoMigrationAndJust`로 확장 또는 신규 테스트 추가.
  - `TestHelpMessageUsesLongRecommendExpansion` 기대 문구 갱신.
- `internal/initcli/render_test.go`
  - custom `MigrationDir`로 렌더링된 `justfile`이 해당 경로를 snapshot 정리 대상으로 쓰는지 확인하는 테스트 추가.
  - `just --dry-run snapshot` smoke를 custom migration dir fixture에도 적용할 수 있음.
- `SPEC.md`
  - `--recommend` 계약과 `just snapshot` 정리 대상 설명을 최신화.
- `README.md`
  - `--recommend` 설명과 generated just command 설명을 최신화.
- `docs/works/justfile-generation/`
  - 이전 open issue였던 custom `--migration-dir` 정리 문제를 resolved로 이동하고 결정 기록 갱신.
- `docs/works/justfile-migration-dir-recommend/`
  - 이번 작업의 학습, 결정, 이슈, 문제 기록.

## 현재 계약 또는 명세 근거

- `SPEC.md`는 `--migration-dir`이 generated Go migration file directory를 지정한다고 정의한다.
- `templates/main.go.tmpl`은 `migratecmd.Config{Dir: "{{.MigrationDir}}"}`를 렌더링하므로 PocketBase migrate command가 생성하는 Go migration file은 `--migration-dir` 경로에 저장된다.
- 현재 `just snapshot`은 migrate command 실행 후 `migrations/*.go`만 정리한다. 따라서 custom `--migration-dir` 사용 시 생성 위치와 정리 위치가 불일치한다.
- 사용자 최신 요청은 이 불일치를 명시적으로 수정하라고 요구한다.
- 사용자 최신 요청은 `-r` flag에 `--just`도 추가하라고 요구한다.
- 현재 `templates/justfile.tmpl` 경로는 이미 요구와 일치한다.

## 재현 경로와 관찰 증상

계획 단계에서 production code는 수정하지 않았고, 현재 baseline `go test ./...`는 통과한다.

예상 재현 경로:

```sh
go run . example.com/custom-migrations --just --migration-dir=internal/migrations
cd custom-migrations
just snapshot -y
```

현재 문제:

- `./pocketbase migrate collections`는 generated app의 `migratecmd.Config{Dir: "internal/migrations"}`에 따라 `internal/migrations`에 migration 파일을 생성한다.
- generated `justfile`은 `migrations` 디렉토리만 확인하고 정리하므로 최신 migration 하나만 유지한다는 목표를 custom migration dir에서 달성하지 못한다.

`--recommend` 관련 현재 문제:

```sh
go run . example.com/recommended --recommend
```

- 현재 `--recommend`은 `--docker --auto-migration`만 켜므로 `justfile`이 생성되지 않는다.

## 원인 가설 또는 확인된 원인

- `snapshot` 문제의 원인은 `templates/justfile.tmpl` 안에서 `migrations` 경로를 hardcoding한 것이다. `renderData.MigrationDir`는 이미 존재하지만 justfile template에서 사용하지 않는다.
- `--recommend` 문제의 원인은 `applyRecommend`가 `cfg.Just`를 설정하지 않는 것이다.
- 문서 불일치의 원인은 이전 구현 당시 `--recommend`과 `snapshot` 정리 대상 계약이 새 요구사항으로 바뀌지 않았기 때문이다.

## 제안 변경 사항

### `templates/justfile.tmpl`

- `snapshot` recipe 내부에 generated migration directory 값을 주입한다.

```bash
migration_dir="{{.MigrationDir}}"
```

- 모든 정리 대상 확인과 `find` 호출을 `migrations`에서 `"$migration_dir"`로 변경한다.

```bash
if [[ ! -d "$migration_dir" ]]; then
    exit 0
fi

find "$migration_dir" -maxdepth 1 -type f -name '*.go' -print | sort
```

- 나머지 `-y`, `--` sentinel, prompt, 최신 timestamp 판정 로직은 유지한다.

### `internal/initcli/cli.go`

- `applyRecommend`에 `cfg.Just = true`를 추가한다.
- `HelpMessage`의 `--recommend` 설명을 `Equivalent to --docker --auto-migration --just.`로 변경한다.

### `internal/initcli/cli_test.go`

- `TestParseRecommendExpandsDockerAutoMigration`를 업데이트한다.
  - 새 기대: `cfg.Docker`, `cfg.AutoMigration`, `cfg.Just`가 모두 true.
- `TestHelpMessageUsesLongRecommendExpansion`를 업데이트한다.
  - `Equivalent to --docker --auto-migration --just.` 포함.
  - 기존 `Equivalent to --docker --auto-migration.`만 단독으로 설명하는 문구는 없어야 함.

### `internal/initcli/render_test.go`

- `TestRenderJustfileUsesConfiguredMigrationDirInSnapshot` 추가.
  - `Config{MigrationDir: "internal/migrations", Just: true}`로 렌더링.
  - generated `justfile`에 `migration_dir="internal/migrations"`가 포함되는지 확인.
  - `find migrations` 또는 `[[ ! -d migrations ]]` 같은 hardcoded 정리 대상이 없는지 확인.
- `TestRenderedJustfileSyntaxSupportsDryRun`를 custom migration dir fixture로 확장하거나 별도 dry-run 테스트를 추가.
  - `just --dry-run snapshot -- -y`가 custom `MigrationDir` 렌더링 상태에서도 parse 가능한지 확인.

### `SPEC.md`

- `--recommend` 설명을 `--docker --auto-migration --just`와 동일하다고 변경.
- `just snapshot` 설명에서 `migrations` 디렉토리를 `--migration-dir` 경로로 변경.
- template 변수 설명에 `templates/justfile.tmpl`에서 `{{.MigrationDir}}`를 사용한다고 명시.

### `README.md`

- `--recommend` 설명을 `--docker --auto-migration --just`로 변경.
- `just snapshot` 설명을 `migrations/` 고정이 아니라 configured migration directory 기준으로 변경.
- examples에서 recommended project가 `justfile`도 생성한다는 점을 자연스럽게 반영한다.

### `docs/works/*`

- `docs/works/justfile-generation/issues.md`의 custom `--migration-dir` open issue를 resolved로 이동한다.
- `docs/works/justfile-generation/decisions.md`에 새 요구로 기존 결정이 변경됐음을 기록한다.
- 새 작업 디렉토리 `docs/works/justfile-migration-dir-recommend/`에 이번 계획, 학습, 결정, 이슈, 문제를 기록한다.

## TDD 계획

1. `TestParseRecommendExpandsDockerAutoMigrationAndJust` 또는 기존 recommend 테스트 업데이트 후 실패를 확인한다.
2. `TestHelpMessageUsesLongRecommendExpansion` 기대 문구를 `--docker --auto-migration --just`로 변경하고 실패를 확인한다.
3. `TestRenderJustfileUsesConfiguredMigrationDirInSnapshot`를 추가하고 실패를 확인한다.
4. 필요하면 `TestRenderedJustfileSyntaxSupportsDryRun`를 custom migration dir로 확장하고 실패를 확인한다.
5. `applyRecommend`, help text, `templates/justfile.tmpl`을 최소 변경해 테스트를 통과시킨다.
6. `SPEC.md`, `README.md`, `docs/works/*`를 구현과 맞게 갱신한다.
7. 리팩터링 필요 여부를 확인하고 테스트를 다시 실행한다.

## 변경 후 기대 동작

- `go run . example.com/app --just --migration-dir=internal/migrations`로 생성된 `justfile`의 `snapshot` recipe는 `internal/migrations`를 정리 대상으로 사용한다.
- default `--migration-dir`인 경우에는 기존처럼 `migrations`를 정리 대상으로 사용한다.
- `go run . example.com/app --recommend`는 `Dockerfile`, `.dockerignore`, auto migration 설정, `justfile`을 모두 생성한다.
- `-r`도 `--recommend`와 동일하게 `--just`를 활성화한다.
- `--recommend`으로 인해 `--docker --just` 조합이 되므로 `.dockerignore`에는 `justfile` 항목이 포함된다.

## 예상 부작용과 호환성 위험

- `--recommend`의 생성 파일 집합이 늘어난다. 기존에는 생성되지 않던 `justfile`이 생성되므로 사용자가 `--recommend`을 쓰는 프로젝트에서 파일 diff가 추가된다.
- help/README/SPEC의 `--recommend` 설명이 변경된다.
- generated `justfile`은 `--migration-dir` 값이 shell 문자열로 들어간다. 현재 `--migration-dir`는 Go import path와 directory path에도 사용되므로, 기존 생성 코드가 처리할 수 없는 문자를 별도로 지원하지 않는다.
- 이전 `docs/works/justfile-generation/`의 "migrations 고정" 결정은 새 요구사항에 의해 변경된다. 변경 이력을 명확히 남겨야 한다.

## 검증 단계

- 단위 테스트:

```sh
go test ./...
```

- 빌드:

```sh
go build ./...
```

- smoke:

```sh
tmpdir=$(mktemp -d)
go build -o "$tmpdir/pb-init" .
cd "$tmpdir"
./pb-init example.com/recommended --recommend --migration-dir=internal/migrations
test -f recommended/justfile
test -f recommended/.dockerignore
grep -F 'justfile' recommended/.dockerignore
grep -F 'migration_dir="internal/migrations"' recommended/justfile
just --justfile recommended/justfile --working-directory recommended --dry-run snapshot -- -y
```

## TODO 체크리스트

- [x] `planner`, `working-docs`, `AGENTS.md` 지침 확인.
- [x] 현재 branch/status 확인.
- [x] 현재 justfile template, render data, `applyRecommend`, tests, SPEC, README 조사.
- [x] baseline `go test ./...` 통과 확인.
- [x] 계획 문서 작성.
- [x] 사용자 계획 승인 받기.
- [x] `main` branch 작업 상태 확인.
- [x] 실패 테스트 추가 및 예상 실패 확인.
- [x] `templates/justfile.tmpl` migration dir templating 구현.
- [x] `--recommend`/`-r`의 `--just` 활성화 구현.
- [x] SPEC/README/work docs 최신화.
- [x] 전체 테스트와 build 실행.
- [x] smoke 검증 실행.
- [x] commit 생성.
- [x] `main` branch 직접 작업으로 merge/rebase 불필요 확인.
