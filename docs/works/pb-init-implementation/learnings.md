# Learnings

## TL;DR

- 현재 저장소는 `SPEC.md`와 템플릿만 있는 초기 상태이며 실행 가능한 Go package가 없다.
- `go run . --help`는 `no Go files in /Users/crmin/workspace/crmin/pb-init`로 실패한다.
- `go test ./...`는 package가 없어 실패한다.
- PocketBase latest는 계획 조사 시점인 2026-07-01 기준 `v0.39.5`이며, `plugins/jsvm`의 등록 API는 `jsvm.MustRegister(app, jsvm.Config{})` 형태다. 구현은 특정 버전에 pin하지 않고 `latest` 설치 결과 기준으로 검증한다.

## 근거

- `SPEC.md`는 `go run github.com/crmin/pb-init [moduleName] [args...]` 실행 계약을 정의한다.
- `templates/main.go.tmpl`은 `--jsvm`일 때 import만 추가하고 호출을 하지 않아 생성 코드에서 unused import 또는 기능 미활성화 문제가 발생할 수 있다.
- `templates/migration_init.go.tmpl`은 `package {{.MigrationDir}}`를 사용하므로 `--migration-dir=internal/migrations` 같은 중첩 경로를 그대로 넣으면 유효한 Go code가 아니다.
- `go list -m -json github.com/pocketbase/pocketbase@latest` 결과 latest는 `v0.39.5`, module Go version은 `1.25.0`이다.
- 1차 subagent 검토에서 명세에 없는 `--migration-dir` 경로 제한과 `moduleName` validation은 사용자 확인 없는 의미 있는 동작 제한으로 지적되어 계획에서 제거했다.
- `MigrationPackage` 변수 추가는 중첩 migration directory 지원을 위한 구현상 필요 사항이므로, 구현 계획에 `SPEC.md` 보완을 포함했다.
- 사용자가 승인 전 추가로 `--migration-dir` 절대 경로와 parent directory reference 거부를 명시했으므로, 해당 제한은 명세 밖 임의 결정이 아니라 계획 승인 대상 계약으로 반영한다.
- Dockerfile에서 `--jsvm` asset directory는 PocketBase 기본 directory인 `pb_migrations`, `pb_hooks`를 대상으로 한다.
- 사용자가 승인 전 추가로 `--migration-dir` current directory reference 거부와 `--pb-version=none` 금지를 명시했으므로, 두 제한은 계획 승인 대상 계약으로 반영한다.
- `--jsvm` flag 자체도 `pb_migrations`, `pb_hooks` 빈 디렉토리를 생성해야 하며, Dockerfile copy 여부와 별개로 적용된다.
- Commit 1 구현에서 parser, 고정 help/error message, stderr routing skeleton을 추가했고 `go test ./...`, `go run . --help`가 통과했다.
- Commit 2 구현에서 현재 디렉토리 Go module 판정, force guard, moduleName 기반 `go mod init`, `go get`, module path 읽기를 추가했고 `go test ./...`가 통과했다.
- `go run . --pb-version=none` smoke에서 pb-init 오류와 help는 stderr로 출력되며 stdout은 비어 있었다. Go tool은 프로그램 exit 1을 감싸 `exit status 1`을 stderr에 추가로 출력한다.
- Commit 3 smoke 중 `--jsvm` generated project는 `go get github.com/pocketbase/pocketbase@latest`만으로 `plugins/jsvm` transitive dependency의 go.sum 항목이 부족해 `go build ./...`가 실패했다. 해결을 위해 `--jsvm`일 때 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}`를 추가 실행한다.
- Commit 3 재검증에서 binary smoke로 `--docker -mj --migration-dir=internal/migrations` 프로젝트 생성, `pb_migrations`/`pb_hooks` 생성, Dockerfile JSVM copy 라인, generated project `go build ./...`를 확인했다.
- Commit 3 current directory mode smoke에서 빈 Go module에 초기화 후 `main.go`, `migrations/init.go`, `.gitignore` 생성과 generated project `go build ./...`를 확인했다.
- Commit 4에서 영어 README를 작성했고 `go test ./...`, `go build ./...`, `go run . --help`, invalid `--migration-dir=.` stderr smoke, moduleName project generation smoke, current directory mode smoke가 모두 통과했다.
- 2026-07-02 baseline `go test ./...`는 `internal/initcli/render_test.go`의 `.dockerignore` binary name newline 기대값에서 실패한다.
- 현재 `moduleName` 경로의 대상 디렉토리가 이미 Go module이어도 `createModuleProject`가 항상 `go mod init`을 실행하므로 `go.mod already exists` 오류가 발생할 수 있다.
- 현재 `PrepareProject`는 PocketBase SDK와 선택적 JSVM plugin `go get` 후 `go mod tidy`를 실행하지 않는다.
- `go mod tidy`는 generated `main.go`가 렌더링된 이후 실행해야 PocketBase import가 유지된다.
- `github.com/fatih/color`의 `Color.EnableColor()`를 사용하면 stdout이 terminal이 아니어도 테스트에서 ANSI foreground color 출력을 안정적으로 확인할 수 있다.
- `moduleName` 대상 디렉토리가 이미 Go module인 경우 `go mod init`을 생략하고 current directory mode의 `--force` 기준을 재사용하면 `go.mod already exists` 오류를 피할 수 있다.
- `.dockerignore.tmpl`과 `.gitignore.tmpl`의 마지막 template action 뒤에 newline을 유지해야 generated binary name이 `\napp\n` 형태로 검증된다.
- ANSI color sequence를 smoke script에서 확인할 때는 `grep` 정규식보다 `grep -F` 고정 문자열 검색을 사용해야 `[`와 `]`를 regex 문자로 오해하지 않는다.
- 2026-07-02 검증에서 `go mod tidy`, `go test ./...`, `go build ./...`, 신규 module 생성 smoke, 기존 target module guard smoke, `--force` smoke, generated project `go build ./...`가 통과했다.

## 재사용 키워드

- pb-init
- remote module run
- PocketBase initializer
- short flag bundle
- migration-dir
- jsvm
- embed templates
- go mod tidy
- fatih/color
- success next steps
