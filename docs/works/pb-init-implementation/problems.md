# Problems

## 문제 정의

- `SPEC.md`는 원격 모듈 실행형 PocketBase 프로젝트 초기화 CLI 계약을 정의하지만, 현재 저장소에는 실행 가능한 Go source package가 없다.
- 현재 `templates/`는 존재하지만 이를 embed하고 렌더링하는 CLI 구현이 없다.
- `moduleName`이 전달된 대상 디렉토리가 이미 Go module이면 현재 구현은 `go mod init`을 다시 실행해 `go.mod already exists` 오류로 종료한다.
- 초기화 성공 후 next step 안내 메시지가 없어 사용자가 `serve`, collection snapshot, superuser 생성 명령을 바로 확인할 수 없다.
- PocketBase SDK 설치 후 generated module에서 `go mod tidy`가 실행되지 않는다.

## 재현 절차

```sh
go run . --help
```

현재 결과:

```text
no Go files in /Users/crmin/workspace/crmin/pb-init
```

```sh
go test ./...
```

현재 결과:

```text
go: warning: "./..." matched no packages
no packages to test
```

2026-07-02 현재 baseline test 결과:

```sh
go test ./...
```

```text
--- FAIL: TestRenderDockerFilesUseCgoAndBinaryName
    render_test.go:66: .dockerignore missing binary name:
FAIL
```

## 원인 후보

- root package에 `main.go`가 없다.
- CLI 인자 파서와 project initialization orchestration 구현이 없다.
- 템플릿 embed 및 렌더링 구현이 없다.
- `createModuleProject`가 대상 디렉토리의 Go module 여부를 확인하지 않고 항상 `go mod init`을 실행한다.
- 초기화 성공 흐름에 `go mod tidy` command 단계가 없다.
- `Run` 성공 경로에 완료 안내 출력이 없다.

## 회귀 방지 수단

- parser, module preparation, rendering 단위 테스트를 추가한다.
- 실제 temp directory에서 `go run` 기반 smoke test와 generated project `go build ./...`를 수행한다.
- `moduleName` 대상 기존 Go module, `go mod tidy` command 순서, color 완료 메시지, 단계별 로그를 단위 테스트로 고정한다.

## 해결 내용 요약

- `moduleName` 대상 디렉토리가 이미 Go module이면 `go mod init`을 생략하고 `go.sum` 또는 root `*.go` 기준으로 `--force`를 요구하도록 변경.
- starter file 렌더링 이후 프로젝트 모듈 디렉토리에서 `go mod tidy`를 실행하도록 변경.
- stdout 단계 로그와 color 완료 안내 메시지를 추가.
- `.dockerignore.tmpl`, `.gitignore.tmpl` 마지막 줄의 trailing newline을 유지해 기존 테스트 실패를 해소.
