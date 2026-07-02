# Learnings

## TL;DR

- 현재 `pb-init`은 템플릿 기반으로 starter file, migration init, 선택적 Docker 파일, `.gitignore`를 렌더링하지만 `justfile` 생성 경로는 없다.
- `just 1.46.0`에서 `[private] default`는 `just --list`와 `just --summary`에 표시되지 않는다.
- `set positional-arguments := true`를 사용하면 shebang recipe 안에서 variadic recipe 인자를 `"$@"`로 받을 수 있고, 공백 포함 인자도 보존된다.
- bash 3.2에서는 `set -u`와 빈 배열 확장 조합에 주의해야 하므로, 빈 배열일 때는 command를 별도 분기로 실행하는 것이 안전하다.
- `just` actual run에서 shebang recipe 본문은 echo되지 않는다. `just --dry-run`에서 본문이 보이는 것은 정상이다.

## 근거

- `internal/initcli/render.go`의 `RenderProject`는 `cfg.Docker`일 때만 `Dockerfile`, `.dockerignore`를 생성하고, `cfg.Just`에 해당하는 분기는 아직 없다.
- `templates/.dockerignore.tmpl`은 현재 `.git`, `.gitignore`, `.github`, `docs/*`, `*.env*`, `*.md`, `*.log`, `pb_data/*`, `pocketbase`, `BinaryName`만 포함한다.
- subagent 1차 검증에서 기존 후보의 `snapshot`은 bash 3.2 `set -u`에서 빈 배열 확장 문제로 실패할 수 있음이 확인됐다.
- subagent 2차 검증에서 수정 후보는 `just 1.46.0`과 `bash 3.2.57` 기준으로 승인 가능하다는 결과를 받았다.
- `snapshot -y -- -y --flag "value with spaces"` 형태에서 `--` 뒤의 `-y`와 공백 포함 값이 PocketBase 인자로 보존됨이 subagent 검증으로 확인됐다.
- `migrations/*.go` 중 numeric timestamp prefix가 가장 큰 파일 하나만 최신 파일로 두고, 나머지 `.go` 파일을 삭제 대상으로 잡는 로직이 subagent 검증으로 확인됐다.
- `AGENTS.md`가 참조하는 `RTK.md`는 현재 작업트리와 상위 3단계 검색에서 발견되지 않았다.
- 구현 전 `go test ./...`는 `Config.Just` 필드와 `templates/justfile.tmpl` 부재로 실패해 신규 테스트가 현재 미구현 계약을 잡는 것을 확인했다.
- `--just` parser/help, `justfile` 렌더링, `.dockerignore` 조건 렌더링 구현 후 `go test ./...`가 통과했다.
- `just` dry-run 테스트는 실제 `go`, `./pocketbase`, `rm` 실행 없이 justfile parse와 recipe argument 전달 문법을 확인한다.
- 후속 요구사항에 따라 `snapshot` 정리 대상은 generated `justfile`의 `migrations` 고정값이 아니라 configured `--migration-dir` 경로로 변경되어야 한다.
- 최종 검증에서 `go test ./...`, `go build ./...`, 임시 프로젝트 `--just --docker` 생성 smoke, `.dockerignore`의 `justfile` 항목 확인, `just --list` recipe 목록 확인, `just snapshot` dry-run syntax 확인이 통과했다.

## 재사용 키워드

- pb-init
- justfile
- just positional-arguments
- shebang recipe
- PocketBase migrate collections
- collection snapshot
- dockerignore
- bash 3.2
