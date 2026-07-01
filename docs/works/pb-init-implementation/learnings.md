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

## 재사용 키워드

- pb-init
- remote module run
- PocketBase initializer
- short flag bundle
- migration-dir
- jsvm
- embed templates
