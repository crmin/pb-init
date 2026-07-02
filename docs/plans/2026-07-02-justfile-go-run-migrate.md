# justfile migrate go run 전환 계획

## 목표와 현재 동작

- 목표: generated `justfile`의 `migrate`와 `snapshot` recipe가 생성 직후 존재하지 않는 `./pocketbase` binary에 의존하지 않도록 `go run . migrate collections`를 사용한다.
- 목표: `snapshot`은 기존처럼 configured `--migration-dir` 경로에서 최신 snapshot 하나만 유지한다.
- 현재 동작:
  - `just migrate`는 `./pocketbase migrate collections "$@"`를 실행한다.
  - `just snapshot`은 `printf 'y\n' | ./pocketbase migrate collections ...`를 실행한다.
  - 생성 직후 프로젝트에는 `./pocketbase` binary가 없으므로 `just migrate`가 `No such file or directory`로 실패한다.

## 관련 파일

- `templates/justfile.tmpl`: migrate/snapshot recipe command 변경.
- `internal/initcli/render_test.go`: generated justfile command 계약 테스트.
- `SPEC.md`, `README.md`: generated just command 설명 갱신.
- `docs/works/justfile-go-run-migrate/`: 작업 기록.

## TDD 계획

1. generated justfile에 `go run . migrate collections`가 포함되고 `./pocketbase migrate collections`가 없는지 테스트를 추가한다.
2. 테스트가 기존 구현에서 실패하는지 확인한다.
3. `templates/justfile.tmpl`의 migrate/snapshot command를 `go run . migrate collections`로 변경한다.
4. 문서와 작업 기록을 갱신한다.
5. `go test ./...`, `go build ./...`, 임시 프로젝트 smoke를 실행한다.

## 기대 동작

- `just migrate [args...]`는 `go run . migrate collections [args...]`를 실행한다.
- `just snapshot [-y] [-- args...]`는 `printf 'y\n' | go run . migrate collections [args...]`를 실행한다.
- 별도 `pocketbase` binary build 없이 생성 직후 `just migrate`/`just snapshot`이 실행 가능하다.

## TODO 체크리스트

- [x] 현재 상태와 재현 원인 확인.
- [x] 실패 테스트 추가 및 실패 확인.
- [x] `templates/justfile.tmpl` 수정.
- [x] SPEC/README/work docs 갱신.
- [x] 테스트, 빌드, smoke 검증.
- [x] commit 생성.
