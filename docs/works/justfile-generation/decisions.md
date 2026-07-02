# Decisions

## Current (Active)

- 계획 우선 진행 - `planner` 지침에 따라 production code는 사용자 승인 전 수정하지 않는다.
- 작업 문서 유지 - `working-docs` 지침에 따라 `docs/plans/`와 `docs/works/justfile-generation/`에 계획, 학습, 결정, 이슈, 문제를 기록한다.
- `--just`는 long flag만 제공 - 사용자 요청에 short flag가 없고 기존 short bundle 계약을 넓히지 않기 위함.
- `--recommend`은 `--just`를 포함하지 않음 - 기존 `--docker --auto-migration` 의미를 유지하기 위함.
- `justfile`은 module root에 lower-case `justfile`로 생성 - 사용자 요청의 파일명을 그대로 따른다.
- `.dockerignore`의 `justfile` 항목은 `--docker --just` 조합에서만 생성 - Docker 파일 생성 계약과 사용자 조건을 동시에 만족하기 위함.
- `snapshot`은 최신 numeric timestamp `.go` 하나만 보존 - 사용자의 "가장 최신 상태만 유지" 요구를 우선한다.
- `snapshot` prompt는 `y/Y`만 진행, `n/N`만 취소 - 사용자 입력 규칙을 명시적으로 따른다.
- `snapshot`은 `--` 뒤 인자를 PocketBase로 보존 - recipe 전용 `-y`와 migrate 인자 전달을 분리하기 위함.
- `upgrade`는 starts-with 규칙을 그대로 적용 - 사용자 요청의 `v` prefix, digit prefix 분기를 semver 검증으로 좁히지 않는다.
- subagent 검증 결과를 계획에 반영 - 사용자 요청대로 justfile 문법과 shell script 유효성을 계획 단계에서 검증한다.
- `snapshot` 정리 대상은 `migrations` 디렉토리로 유지 - 사용자 요구와 승인된 계획이 `migrations` 디렉토리를 명시했으므로 custom `--migration-dir` 정리까지 확장하지 않는다.

## Change Log

### 2026-07-02

- Changed: `--just` flag와 generated `justfile` 추가 작업을 새 계획 범위로 정의.
- Reason: 사용자가 실행 옵션과 just command 동작을 신규 요구사항으로 제시함.

- Changed: `justfile` 후보를 `set positional-arguments := true`와 bash shebang recipe 기반으로 결정.
- Reason: variadic 인자의 공백 보존과 bash 환경 기준 실행을 동시에 만족해야 함.

- Changed: subagent 1차 검증 결과에 따라 `snapshot` 빈 배열 처리와 삭제 범위를 수정.
- Reason: bash 3.2 `set -u` 빈 배열 실패와 최신 snapshot 하나만 보존하는 정책의 불명확성을 제거하기 위함.

- Changed: `snapshot`에서 `--` sentinel 뒤 인자는 PocketBase 인자로 보존하도록 결정.
- Reason: recipe 전용 `-y` flag와 migrate command 인자 전달 충돌을 줄이기 위함.

- Changed: subagent 2차 검증 APPROVE 결과를 기록.
- Reason: 수정된 justfile 후보가 just 1.46.0, bash 3.2.57 기준으로 문법과 주요 shell 로직을 충족함을 확인했기 때문.

- Changed: 승인 후 `codex/justfile-generation` 브랜치를 생성하고 실패 테스트를 추가.
- Reason: planner TDD 절차에 따라 구현 전 신규 계약이 실패하는지 확인하기 위함.

- Changed: `--just` parser/help, `templates/justfile.tmpl`, `.dockerignore` 조건 렌더링을 구현.
- Reason: 사용자 요구의 generated justfile과 Docker ignore 조건을 코드 계약으로 반영하기 위함.

- Changed: `SPEC.md`, `README.md`에 `--just`와 generated just command 계약을 반영.
- Reason: 프로젝트 문서는 현재 코드 상태를 항상 반영해야 하기 때문.

- Changed: 구현, 테스트, 명세, README, 작업 문서를 하나의 기능 커밋으로 정리하고 `main` merge를 진행.
- Reason: `--just` 생성 옵션은 하나의 원자적 기능 단위이며, 관련 문서는 코드 변경과 같은 커밋에 포함되어야 하기 때문.
