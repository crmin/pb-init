# PocketBase SDK tidy 및 초기화 출력 개선 계획

## 목표와 현재 동작

- 목표: PocketBase SDK 설치 후 generated module에서 `go mod tidy`를 실행하고, 초기화 진행 단계 로그와 완료 후 안내 메시지를 추가한다.
- 목표: `moduleName`이 전달된 경우에도 대상 디렉토리가 이미 Go module인지 먼저 확인해, 기존 current directory mode와 같은 `--force` 보호 흐름을 적용한다.
- 목표: 새 동작을 `SPEC.md`와 `README.md`에 반영하되, `SPEC.md`는 기존 문체와 구조를 유지하면서 최소 수정한다.
- 현재 동작:
  - `PrepareProject`는 `go get github.com/pocketbase/pocketbase@{version}`와 선택적 JSVM plugin `go get`까지만 실행하고 `go mod tidy`를 실행하지 않는다.
  - `Run`은 성공 시 완료 메시지를 출력하지 않는다.
  - 진행 단계 로그는 없다. current directory mode에서 `--force`가 필요한 경우에만 stdout 경고가 출력된다.
  - `moduleName`이 전달되면 `createModuleProject`가 대상 디렉토리를 만든 뒤 항상 `go mod init`을 실행한다. 대상에 이미 `go.mod`가 있으면 Go tool의 `go.mod already exists` 오류가 그대로 출력된다.
  - 2026-07-02 기준 baseline `go test ./...`는 `TestRenderDockerFilesUseCgoAndBinaryName`에서 `.dockerignore`의 마지막 binary name 뒤 newline 기대값 때문에 실패한다.

## 관련 파일과 코드 위치

- `internal/initcli/project.go`
  - `PrepareProject`: 프로젝트 해석, PocketBase SDK 설치, JSVM plugin 설치, module path 읽기 흐름.
  - `resolveProject`: `moduleName` 유무에 따른 current module 또는 new module 분기.
  - `createModuleProject`: `moduleName` 대상 디렉토리 생성과 `go mod init` 실행.
  - `currentModuleRequiresForce`: `go.sum` 또는 root `*.go` 존재 시 `--force` 필요 여부 판단.
  - `runGoGetPackage`, `runCommand`: 외부 Go command 실행과 실패 출력 전달.
- `internal/initcli/cli.go`
  - `Run`: stdout/stderr wiring, `PrepareProject`, `RenderProject` 호출, 성공 종료 흐름.
  - `Env`, `CommandRunner`: 테스트 가능한 입출력과 외부 command 의존성.
- `internal/initcli/project_test.go`
  - current module force guard, `moduleName` 생성, `go get`, JSVM plugin command 호출 테스트.
- `internal/initcli/cli_test.go`
  - stderr/stdout 출력 경로와 인자 파싱 테스트.
- `internal/initcli/render_test.go`
  - 현재 실패 중인 `.dockerignore` newline 기대 테스트.
- `go.mod`
  - 완료 메시지 색상 출력을 위해 `github.com/fatih/color` 의존성 추가 예정.
- `SPEC.md`, `README.md`
  - 새 실행 단계, 출력 메시지, `moduleName` 대상 기존 Go module 처리 계약을 반영할 문서.
- `docs/works/pb-init-implementation/`
  - 작업 중 확인한 학습, 결정, 문제, 이슈 기록.

## 현재 계약 또는 명세 근거

- `SPEC.md`는 `moduleName`이 있으면 마지막 경로 요소로 하위 디렉토리를 생성하고 `go mod init <moduleName>`을 실행한다고 정의한다.
- `SPEC.md`는 `moduleName`이 없고 현재 디렉토리가 Go module일 때, `go.sum` 또는 root `*.go`가 있으면 `--force`를 요구한다고 정의한다.
- `SPEC.md`는 외부 명령 실패 시 command output을 stderr에 그대로 출력하고 exit code 1로 종료한다고 정의한다.
- `SPEC.md`는 `--force` 경고 메시지는 오류가 아니므로 stdout으로 출력한다고 정의한다.
- 사용자 추가 요구사항은 SDK 설치 후 `go mod tidy`, 단계별 로깅, 색상 있는 성공 안내 메시지, `moduleName` 대상 기존 Go module guard, `SPEC.md` 최소 최신화를 새 계약으로 추가한다.

## 재현 경로와 관찰 증상

- `moduleName` 대상이 이미 Go module인 경우:

```sh
go run github.com/crmin/pb-init@latest github.com/crmin/pb-test -r
```

현재 관찰 증상:

```text
go: /Users/crmin/workspace/scratch/pb-test/go.mod already exists
exit status 1
```

- baseline test:

```sh
go test ./...
```

현재 관찰 증상:

```text
--- FAIL: TestRenderDockerFilesUseCgoAndBinaryName
    render_test.go:66: .dockerignore missing binary name:
...
FAIL
```

## 원인 가설 또는 확인된 원인

- `moduleName` 대상 기존 Go module 오류는 `createModuleProject`가 대상 디렉토리의 `go.mod`를 확인하지 않고 항상 `go mod init`을 실행하기 때문에 발생한다.
- `go mod tidy` 누락은 초기화 성공 흐름에 tidy command가 없기 때문이다.
- 성공 안내 메시지와 단계 로그 누락은 `Run`, `PrepareProject`, `RenderProject` 주변에 사용자-facing 출력 함수가 없기 때문이다.
- baseline test 실패는 `.dockerignore` 템플릿의 마지막 줄이 trailing newline 없이 렌더링되어 테스트 기대값 `\napp\n`과 맞지 않는 문제로 보인다.

## 제안 변경 사항

### `internal/initcli/project.go`

- `Project`에 성공 안내 메시지 분기용 정보를 추가한다.
  - `Dir`: absolute module directory로 정규화.
  - `RelativeDir`: command 실행 디렉토리 기준 module directory 상대 경로. `moduleName`이 전달된 경우 완료 메시지의 `cd {module relative path}`에 사용한다.
  - `CreatedFromModuleName` 또는 동등한 bool: 완료 메시지에 `Go to module directory` 섹션이 필요한지 판단한다.
- `PrepareProject`에서 프로젝트 준비와 SDK 설치 단계 로그를 stdout에 출력한다.
  - module directory 확인 또는 생성.
  - `go mod init` 실행.
  - PocketBase SDK 설치.
  - JSVM plugin dependency 설치.
  - starter file render 단계와 `go mod tidy` 단계는 `Run`에서 출력한다.
- `moduleName`이 전달된 대상 디렉토리가 이미 Go module이면 `go mod init`을 다시 실행하지 않는다.
  - 대상이 Go module이고 `go.sum` 또는 root `*.go`가 있으면 current directory mode와 같은 `--force` guard를 적용한다.
  - `--force`가 없으면 기존 `forceRequiredMessage`를 stderr로 출력하고 종료한다.
  - `--force`가 있으면 기존 `forceProceedMessage`를 stdout으로 출력하고 계속 진행한다.
  - 대상이 Go module이지만 `go.sum`과 root `*.go`가 없으면 current directory mode와 같이 추가 경고 없이 계속 진행한다.
  - 대상 디렉토리가 없으면 기존처럼 생성하고 `go mod init <moduleName>`을 실행한다.
  - 대상 디렉토리가 있지만 Go module이 아니면 기존처럼 `go mod init <moduleName>`을 실행한다.

### `internal/initcli/cli.go`

- `github.com/fatih/color`를 사용해 완료 안내 메시지의 지정된 토큰만 색상 처리한다.
  - `{module abs path}`, `go run . serve`, `go run . migrate collections`, `go run . superuser create`는 cyan.
  - `<user_email>`, `<user_password>`는 magenta.
  - `moduleName`이 전달된 경우 `cd {module relative path}` 전체를 cyan.
- `Run`에서 `RenderProject` 성공 후 완료 메시지를 출력한다.
  - `moduleName`이 없었던 current module mode:

```text
PocketBase project initialized successfully: {module abs path}

Start the server:
    go run . serve

Create a collection snapshot:
    go run . migrate collections

Create a superuser:
    go run . superuser create <user_email> <user_password>
```

  - `moduleName`이 전달된 mode:

```text
PocketBase project initialized successfully: {module abs path}

Go to module directory:
    cd {module relative path}

Start the server:
    go run . serve

Create a collection snapshot:
    go run . migrate collections

Create a superuser:
    go run . superuser create <user_email> <user_password>
```

- 기존 parse/help/error stdout/stderr 계약은 유지한다.
- PocketBase SDK 설치와 파일 렌더링이 끝난 뒤 `go mod tidy`를 한 번 실행한다.
  - `--jsvm`이 있으면 base SDK와 JSVM plugin `go get`, starter file render 이후 tidy를 실행한다.
  - tidy는 생성된 `main.go`의 import를 Go tool이 볼 수 있는 상태에서 실행되어야 하므로 렌더링 뒤에 둔다.
  - 실패 시 기존 외부 command 실패와 동일하게 command output을 stderr로 전달하고 exit code 1로 종료한다.

### `internal/initcli/*_test.go`

- command 호출 순서 테스트를 업데이트한다.
  - current module mode: `go get pocketbase`, 선택적 `go get jsvm`, 파일 렌더링 후 `go mod tidy`.
  - `moduleName` 신규 생성: `go mod init`, `go get pocketbase`, 선택적 `go get jsvm`, 파일 렌더링 후 `go mod tidy`.
- `moduleName` 대상 기존 Go module guard 테스트를 추가한다.
  - `TestModuleNameExistingModuleWithGoSumRequiresForce`
  - `TestModuleNameExistingModuleWithGoFilesRequiresForce`
  - `TestModuleNameExistingModuleWithForcePrintsWarningAndSkipsGoModInit`
  - `TestModuleNameExistingEmptyModuleSkipsGoModInit`
- 성공 메시지 테스트를 추가한다.
  - `TestCurrentModuleSuccessPrintsColoredNextSteps`
  - `TestModuleNameSuccessPrintsCdStepWithRelativePath`
  - 테스트에서는 ANSI escape를 직접 확인하거나 색상 비활성화 상태를 통제해 `fatih/color` 사용 결과를 안정적으로 검증한다.
- 로깅 테스트를 추가한다.
  - `TestRunPrintsStepLogsInOrder`
  - 실패 시 완료 메시지가 출력되지 않는지 확인한다.
- baseline 실패를 정리한다.
  - `.dockerignore` 또는 `.gitignore` 템플릿의 trailing newline을 유지해 기존 테스트 기대와 POSIX 텍스트 파일 관례를 만족시킨다.

### `go.mod`와 `go.sum`

- `github.com/fatih/color`를 직접 의존성으로 추가한다.
- repository 자체에서도 `go mod tidy`를 실행해 `go.mod`와 `go.sum`을 최신화한다.

### `SPEC.md`

- 기존 문단을 가능한 한 유지하고 새 내용만 최소 삽입한다.
- `moduleName`이 전달된 경우 대상 디렉토리가 이미 Go module이면 current directory mode와 같은 force guard를 적용한다는 내용을 positional argument 동작 설명에 추가한다.
- `동작` section의 PocketBase SDK 설치 단계 뒤에 `go mod tidy` 실행을 추가한다.
- Output Streams 또는 `동작` section에 단계별 로그와 성공 안내 메시지가 stdout으로 출력된다는 계약을 추가한다.
- 색상 출력은 `github.com/fatih/color`를 사용하며, 지정된 토큰의 foreground color를 명시한다.
- 외부 명령 실패 출력 계약은 유지한다.

### `README.md`

- Quick Start와 current directory 설명에 `go mod tidy`, 단계별 로그, 완료 후 next steps를 반영한다.
- `moduleName` 대상 디렉토리가 이미 Go module인 경우 `--force`가 필요할 수 있음을 설명한다.
- README는 사용자 문서이므로 영어 유지.

### `docs/works/pb-init-implementation/`

- 이번 요구사항, baseline test 실패, 구현 결정과 검증 결과를 `learnings.md`, `decisions.md`, `problems.md`, `issues.md`에 기록한다.
- 작업 완료 시 checklist 결과와 검증 결과를 반영한다.

## TDD 계획

1. `TestModuleNameExistingModuleWithGoSumRequiresForce`를 추가하고 실패를 확인한다.
2. `TestModuleNameExistingModuleWithForcePrintsWarningAndSkipsGoModInit`를 추가하고 실패를 확인한다.
3. `TestCurrentModuleRunsGoModTidyAfterRendering` 또는 기존 command order 테스트 업데이트 후 실패를 확인한다.
4. `TestJSVMRunsGoModTidyAfterRendering`을 추가 또는 기존 JSVM command order 테스트 업데이트 후 실패를 확인한다.
5. `TestCurrentModuleSuccessPrintsColoredNextSteps`와 `TestModuleNameSuccessPrintsCdStepWithRelativePath`를 추가하고 실패를 확인한다.
6. `TestRunPrintsStepLogsInOrder`를 추가하고 실패를 확인한다.
7. 구현을 최소 범위로 추가해 테스트를 통과시킨다.
8. 기존 baseline `.dockerignore` newline 실패를 정리하고 전체 테스트를 녹색으로 만든다.
9. 리팩터링 후 테스트를 다시 실행한다.

## 변경 후 기대 동작

- 신규 `moduleName` 실행은 기존처럼 하위 디렉토리를 만들고 `go mod init`, SDK 설치, 선택적 JSVM plugin 설치, 파일 렌더링, `go mod tidy`, 완료 안내를 순서대로 수행한다.
- current directory mode는 기존 force guard를 유지하면서 SDK 설치, 파일 렌더링, `go mod tidy`, 완료 안내를 추가로 수행한다.
- `moduleName` 대상 디렉토리가 이미 Go module이면 더 이상 `go mod init`으로 실패하지 않고, current directory mode와 같은 기준으로 `--force` 필요 여부를 판단한다.
- 성공 메시지는 실행 mode에 따라 `cd {module relative path}` 섹션 포함 여부가 달라진다.
- 지정된 command/path placeholder는 `fatih/color` 기반 cyan/magenta로 출력된다.
- 외부 Go command가 실패하면 기존처럼 command output만 stderr로 전달되고 완료 메시지는 출력되지 않는다.

## 예상 부작용과 호환성 위험

- stdout 출력이 늘어나므로 기존에 성공 시 stdout이 비어 있다고 가정한 테스트와 사용 스크립트는 영향을 받을 수 있다.
- `go mod tidy`는 generated module의 `go.mod`와 `go.sum`을 정리하므로 기존 `go get` 직후보다 require 목록이 달라질 수 있다. 이는 Go toolchain의 일반 동작이지만 diff가 생길 수 있다.
- `github.com/fatih/color` 의존성이 추가되어 repository `go.mod`와 새 `go.sum`이 변경된다.
- `fatih/color`는 환경에 따라 색상 비활성화가 발생할 수 있으므로, 구현 시 사용자 요구의 color 출력과 테스트 안정성을 함께 확인해야 한다.
- `moduleName` 대상 기존 Go module에서 `go.sum`과 root `*.go`가 없는 경우에는 current directory mode와 동일하게 force 없이 진행한다. 이는 "moduleName 없이 실행될 때와 동일하게"라는 요구에 맞춘 해석이다.

## 검증 단계

- 단위 테스트:

```sh
go test ./...
```

- 빌드:

```sh
go build ./...
```

- 수동 smoke:

```sh
tmpdir=$(mktemp -d)
cd "$tmpdir"
go run /Users/crmin/workspace/crmin/pb-init github.com/crmin/pb-test -r
```

- existing target module guard smoke:

```sh
tmpdir=$(mktemp -d)
mkdir -p "$tmpdir/pb-test"
cd "$tmpdir/pb-test"
go mod init github.com/crmin/pb-test
touch go.sum
cd "$tmpdir"
go run /Users/crmin/workspace/crmin/pb-init github.com/crmin/pb-test -r
go run /Users/crmin/workspace/crmin/pb-init github.com/crmin/pb-test -r --force
```

- generated project build smoke:

```sh
cd "$tmpdir/pb-test"
go build ./...
```

- 출력 확인:
  - 성공 메시지에 absolute module path가 포함되는지 확인한다.
  - `moduleName` mode에서 `cd pb-test`가 출력되는지 확인한다.
  - 지정된 토큰에 ANSI color sequence가 포함되는지 확인한다.
  - 에러 상황에서 완료 메시지가 출력되지 않는지 확인한다.

## TODO

- [x] 현재 구현, 명세, README, 테스트 구조 조사
- [x] baseline `go test ./...` 결과 확인
- [x] 계획 문서 작성
- [x] 사용자 승인 받기
- [x] 실패하는 테스트 먼저 추가 또는 업데이트
- [x] `moduleName` 대상 기존 Go module guard 구현
- [x] SDK 설치 후 `go mod tidy` 구현
- [x] 단계별 로그와 color 완료 메시지 구현
- [x] baseline `.dockerignore` newline 테스트 실패 정리
- [x] `SPEC.md` 최소 수정으로 최신화
- [x] `README.md` 최신화
- [x] works 문서에 구현/검증 결과 반영
- [x] `go mod tidy`, `go test ./...`, `go build ./...` 실행
- [x] 수동 smoke 검증
- [x] 원자 단위 commit 수행
