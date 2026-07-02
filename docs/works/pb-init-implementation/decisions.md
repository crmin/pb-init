# Decisions

## Current (Active)

- 계획 우선 진행 - `planner` 지침에 따라 production code는 사용자 승인 전 수정하지 않는다.
- 작업 문서 유지 - `working-docs` 지침에 따라 `docs/plans/`와 `docs/works/pb-init-implementation/`에 계획, 학습, 결정, 이슈, 문제를 기록한다.
- CLI 출력 문구 고정 - 사용자 요청에 따라 help/error/warning message는 계획 단계에서 확정하고 테스트로 고정한다.
- commit은 원자 단위 직후 수행 - 사용자 요청과 planner commit policy에 따라 각 작업 단위의 검증 직후 즉시 commit한다.
- README는 영어로 작성 - `go run github.com/crmin/pb-init` 사용자 대상 문서이므로 요청대로 영어 README를 작성한다.
- 명세 밖 moduleName 선제 제한 제거 - `moduleName`에는 명세에 없는 별도 validation을 추가하지 않는다.
- `--migration-dir` 경로 제한 추가 - 사용자 추가 요청에 따라 절대 경로와 `..` path component를 거부하고 고정 오류 메시지를 stderr에 출력한다.
- `--migration-dir` current directory reference 금지 - 사용자 추가 요청에 따라 `.` path component도 거부하고 고정 오류 메시지를 stderr에 출력한다.
- `--pb-version=none` 금지 - 사용자 추가 요청에 따라 정확히 `none` 값은 `go get`에 전달하지 않고 고정 오류 메시지로 종료한다.
- 에러 출력 채널 고정 - `--help`는 stdout, 에러 메시지와 외부 명령 실패 output은 stderr로 출력한다. `go get` 실패 출력은 기존 `SPEC.md`의 stdout 계약을 stderr로 변경하고, `go mod init` 실패 출력은 새 stderr 계약으로 추가한다.
- JSVM 기본 디렉토리 생성 - `--jsvm` 전달 시 PocketBase project module directory에 `pb_migrations`, `pb_hooks` 빈 디렉토리를 생성한다.
- JSVM plugin dependency 보강 - `--jsvm` 전달 시 generated project build를 위해 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}`를 추가 실행한다.
- JSVM Docker asset 포함 - `--jsvm --docker` 조합에서 Dockerfile final stage에 `pb_migrations`, `pb_hooks`를 copy한다.
- `MigrationPackage`는 SPEC 보완과 함께 추가 - 중첩 migration directory를 유효한 Go package로 생성하기 위해 template variable을 추가하되 같은 commit에서 `SPEC.md`를 보완한다.
- SDK 설치와 파일 렌더링 후 tidy 실행 - 사용자 요청에 따라 generated module에서 PocketBase SDK와 선택적 JSVM plugin 설치 후 `go mod tidy`를 실행하되, generated imports가 제거되지 않도록 파일 렌더링 뒤에 실행한다.
- `moduleName` 대상 기존 module guard - `moduleName`이 전달된 경우에도 대상 디렉토리가 Go module이면 current directory mode와 같은 `--force` 기준을 적용한다.
- 성공 안내 색상 출력 - 완료 메시지의 지정 token은 `github.com/fatih/color`로 cyan 또는 magenta foreground를 적용한다.
- SPEC 최소 최신화 - 새 동작은 `SPEC.md` 기존 문체와 구조를 유지하며 필요한 문장만 삽입한다.

## Change Log

### 2026-07-01

- Changed: pb-init 구현 계획의 문서 구조와 초기 구현 결정을 기록.
- Reason: 명세 기반 구현 전 사용자의 계획 승인과 subagent 검토가 필요함.

- Changed: 1차 subagent reject 결과에 따라 `moduleName` validation과 `--migration-dir` 경로 제한을 계획에서 제거하고, `MigrationPackage` 추가를 `SPEC.md` 보완 예정 항목으로 분리.
- Reason: 명세에 없는 의미 있는 동작 제한과 명세 불일치를 제거하기 위함.

- Changed: 2차 subagent 검토 APPROVE 결과를 기록.
- Reason: 사용자 승인 요청 전 명세 커버리지 검토가 완료되었음을 남기기 위함.

- Changed: 사용자 추가 요청에 따라 `--migration-dir` 절대 경로와 parent directory reference 거부, stderr 출력, Dockerfile JSVM asset copy, help/error 문구 수정을 계획에 반영.
- Reason: 사용자가 계획 승인 전 계약 변경을 명시했고, 정적 메시지와 Dockerfile 변경 방식을 계획 단계에서 고정해야 함.

- Changed: 3차 subagent reject 결과에 따라 `go get` 실패 출력은 기존 `SPEC.md` stdout 계약을 stderr로 변경하는 항목이고, `go mod init` 실패 출력은 새 stderr 계약 추가 항목임을 계획에 명시.
- Reason: 사용자 추가 요청을 반영하는 과정에서 기존 명세와의 차이를 숨기지 않고 명확한 계약 변경으로 다루기 위함.

- Changed: 4차 subagent 검토 APPROVE 결과를 기록.
- Reason: 수정된 계획이 사용자 추가 요청과 명세 커버리지를 충족함을 확인했기 때문.

- Changed: 사용자 추가 요청에 따라 `--pb-version=none` 금지, `--migration-dir` current directory reference 금지, `--jsvm` 시 `pb_migrations`와 `pb_hooks` 빈 디렉토리 생성 동작을 계획에 반영.
- Reason: 사용자가 계획 승인 전 추가 계약 변경을 명시했고, error message와 생성 동작을 계획 단계에서 고정해야 함.

- Changed: 5차 subagent 검토 APPROVE 결과를 기록.
- Reason: 최신 계획이 사용자 추가 요청과 명세 커버리지를 충족함을 확인했기 때문.

- Changed: Commit 1 범위로 parser, 고정 메시지, stderr 출력 skeleton, 관련 `SPEC.md` 계약 보완을 구현.
- Reason: 승인된 계획의 첫 번째 원자 작업 단위 완료.

- Changed: Commit 2 범위로 Go module 판정, force guard, `go mod init`, `go get`, module path 읽기 구현.
- Reason: 승인된 계획의 두 번째 원자 작업 단위 완료.

- Changed: Commit 3 smoke 실패 결과에 따라 `--jsvm` 시 jsvm plugin dependency를 추가 `go get`하도록 계획과 구현 범위를 보완.
- Reason: generated project가 `--jsvm` import를 포함할 때 즉시 `go build ./...`가 통과해야 하기 때문.

- Changed: Commit 3 범위로 템플릿 렌더링, JSVM 디렉토리 생성, Dockerfile JSVM asset copy, generated project build 검증 완료.
- Reason: 승인된 계획의 세 번째 원자 작업 단위 완료.

- Changed: Commit 4 범위로 영어 README 작성과 최종 검증 완료.
- Reason: 승인된 계획의 네 번째 원자 작업 단위 완료.

### 2026-07-02

- Changed: PocketBase SDK 설치 후 `go mod tidy`, 단계별 로그, color 완료 메시지, `moduleName` 대상 기존 Go module guard를 새 계획 범위로 추가.
- Reason: 사용자가 새 초기화 실행 흐름과 출력 계약을 요청했고, 기존 구현은 `go mod init` 중복 실행과 성공 안내 누락이 있음.

- Changed: `SPEC.md` 업데이트는 기존 작성 내용 수정을 최소화하고 새 내용도 기존 문체와 통일되게 작성하는 방식으로 결정.
- Reason: 사용자가 명세 최신화를 추가 요청하면서 기존 문체와 최소 수정 원칙을 명시함.

- Changed: `go mod tidy` 실행 순서를 SDK 설치 직후가 아니라 starter file 렌더링 이후로 보정.
- Reason: 렌더링 전 tidy를 실행하면 아직 generated `main.go` import가 없어 `go get`으로 추가한 PocketBase 의존성이 제거될 수 있기 때문.

- Changed: `github.com/fatih/color`는 완료 메시지에서 `EnableColor()`를 적용해 지정 token에 ANSI color를 강제로 출력하도록 구현.
- Reason: stdout이 terminal이 아닌 테스트/redirect 환경에서도 사용자 요구의 foreground color 계약을 일관되게 검증하기 위함.

- Changed: baseline `.dockerignore`/`.gitignore` 템플릿 마지막 줄에 trailing newline을 유지.
- Reason: generated ignore 파일을 정상 텍스트 파일 형태로 만들고 기존 `.dockerignore` binary name 테스트 실패를 해소하기 위함.
