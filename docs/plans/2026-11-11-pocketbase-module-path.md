# PocketBase 모듈 경로 전환 계획

## 목표와 현재 상태

2026-11-11에 계획을 승인받아 구현과 검증을 마쳤습니다. 사용자의 인용문은 이전 경로와 새 경로가 같습니다. 저장소와 공식 이전 안내를 근거로 `github.com/pocketbase/pocketbase`에서 `ganigeorgiev.com/pocketbase/pocketbase`로 전환했습니다. Go import와 모듈 경로에는 `https://`를 붙이지 않습니다.

변경 전에는 `PrepareProject`, `templates/main.go.tmpl`, 생성된 `justfile`의 `upgrade`가 GitHub 경로를 고정해서 사용했습니다. 설치 명령과 생성 코드, 업그레이드 명령을 함께 변경했습니다.

`SPEC.md`의 동작 1번과 justfile 계약에 경로 선택 기준을 반영했습니다. `--pb-version`은 기본값 `latest`와 `v0.39.5` 같은 명시적 버전을 지원합니다. 루트 `go.mod`에는 PocketBase 직접 의존성이 없습니다.

공식 이전 근거는 [PocketBase 이전 공지](https://github.com/pocketbase/pocketbase/discussions/7873)입니다. Go 모듈 프록시에서 신규 경로의 v0.39.5와 v0.40.6은 기존 GitHub module을 선언하고, v0.41.0과 v0.42.0은 신규 module을 선언함을 확인했습니다. 신규 경로 latest는 v0.42.0이며 GitHub latest는 v0.40.6입니다. [공식 Go 안내](https://pocketbase.io/docs/go-overview/)도 신규 import를 사용합니다.

## 파일별 변경

- `internal/initcli/project.go`의 `PrepareProject`, `runGoGet`에서 SDK와 JSVM 설치 경로 및 진행 로그를 변경합니다.
- `internal/initcli/cli.go`의 `--help` 설치 예시를 변경합니다. 모듈 조회 명령은 JSON을 stdout에서 읽고 실패 시 Go 진단을 stderr에 전달할 수 있게 합니다.
- `templates/main.go.tmpl`의 PocketBase, JSVM, migratecmd import를 변경합니다.
- `templates/justfile.tmpl`의 `upgrade`가 사용하는 모든 설치 경로를 변경합니다.
- `internal/initcli/project_test.go`에서 설치 명령과 로그의 기대값을 갱신합니다.
- `internal/initcli/cli_test.go`에서 도움말에 신규 경로가 표시되는지 확인합니다.
- `internal/initcli/render_test.go`에서 실제 생성된 Go import와 upgrade 명령을 확인합니다.
- `SPEC.md`에 설치와 upgrade 경로 및 확인된 버전 호환 계약을 반영합니다.
- `README.md`의 설치 경로와 버전 예시를 최종 계약에 맞춥니다. 다른 독자를 위한 변경 문장은 Humanize Korean 스킬로 윤문합니다.
- 이 계획과 `docs/works/pocketbase-module-path/`에 조사 결과, 결정, 검증 결과를 기록합니다.

## 호환성 결정

기본 또는 명시적 `latest`는 신규 경로에서 조회합니다. 버전, 태그, 리비전 쿼리는 신규 경로부터 조회하며, 그 경로에서 버전을 찾지 못하면 GitHub 경로에서도 조회합니다. v0.39.5와 v0.40 쿼리는 기존 경로, v0.41.0-rc.1은 신규 경로로 판별합니다. `patch`와 `upgrade`는 현재 대상 모듈의 `go.mod`에서 PocketBase 요구 경로를 확인해 그 경로로 조회합니다. 두 경로가 모두 있거나 아직 요구가 없으면 신규 경로를 사용합니다.

모든 쿼리에서 `go list -m -json <조회 경로>@<query>`의 stdout을 읽습니다. JSON의 `GoMod` 파일에서 module 선언을 읽고 `Version`으로 설치 버전을 고정합니다. 조회 실패와 예상하지 못한 module 선언은 SDK 설치 전에 오류로 처리합니다. 선택한 경로와 확정 버전을 SDK, JSVM 설치 및 로그에서 함께 사용합니다. 경로는 `Project`와 `renderData`를 통해 Go 템플릿과 justfile에 전달합니다. 다른 Go 명령은 기존의 출력 및 오류 전달 방식을 유지합니다.

기존 프로젝트에서 `--force`로 신규 경로를 선택하기 전에 사용자 Go 코드와 마이그레이션의 import도 전환해야 합니다. 이전 경로를 import한 마이그레이션은 별도 레지스트리에 등록되어 새 앱에 적용되지 않습니다. 기존 import를 유지하려면 이전 경로를 선언한 버전을 명시하도록 README에 안내합니다. 생성된 justfile의 `upgrade`는 생성 시 선택한 경로의 버전을 설치합니다. 두 모듈 경로를 오가는 전환은 Go import와 의존성을 함께 바꾸는 작업입니다.

## TDD와 검증

1. 기존 `TestJSVMRunsPluginGoGetForTransitiveDependencies`, `TestRunPrintsStepLogsInOrder`, `TestCurrentModuleWithoutGoSumOrGoFilesUsesCurrentDirectory` 기대값을 수정하고 실패를 확인합니다.
2. `TestRenderMainTemplateUsesModulePathMigrationDirJSVMAndAutoMigration`을 보강해 기본 import, migratecmd, JSVM 경로를 확인합니다. JSVM 비활성 생성도 검증합니다.
3. `TestRenderedJustfileUpgradeUsesPocketBaseModulePath`를 추가해 버전 생략, `latest`, 숫자 버전, v 접두사 버전의 실제 명령을 fake go로 확인합니다.
4. 최소 코드 변경을 적용하고 해당 테스트를 다시 실행합니다. 경로 소유권을 검토해 불필요한 중복과 추상화를 정리합니다.
5. `go test ./...`, `go vet ./...`, `go build ./...`를 실행합니다.
6. 임시 디렉토리에서 빌드된 pb-init으로 기본 프로젝트와 `--jsvm --recommend` 프로젝트를 생성합니다. 각 프로젝트에서 `go build ./...`를 실행해 실제 의존성 해석을 확인합니다.
7. `TestPocketBaseModulePathForVersion`에서 v0.39.5, v0.40, v0.41.0-rc.1, v0.41.0, latest 및 조회 오류를 검증합니다. 별도 케이스에서 기존 GitHub 요구에 대한 `patch`와 `upgrade`, GitHub 경로 재조회를 확인합니다. 이전 버전과 최신 버전 생성물을 실제 빌드하고 기존 모듈의 `--force` 경로를 검증합니다. 실패 시 동일 명령 반복 전에 원인을 구분합니다.
8. 코드, 템플릿, 현재 사용 문서의 경로를 검색해 누락을 확인합니다. 검증 결과와 실행 환경을 작업 기록에 남깁니다.

## 기대 결과와 위험

신규 프로젝트 생성과 업그레이드가 공식 신규 모듈 경로를 사용합니다. 네트워크 또는 Go proxy 상태가 실제 생성 프로젝트 검증에 영향을 줄 수 있습니다. 신규 모듈의 Go 버전 요구사항과 API 호환성을 실제 빌드로 확인합니다.

## Commit Plan

원자적 작업은 설치, 생성 import, upgrade 경로의 동시 전환입니다. 위 코드와 테스트, SPEC, README, 계획 및 작업 기록을 함께 포함합니다. 전체 테스트와 빌드, 생성 프로젝트 검증을 마친 시점에 `fix(init): PocketBase 신규 모듈 경로 전환`으로 커밋합니다.

## TODO

- [x] 저장소 계약과 변경 지점 조사
- [x] 공식 신규 module 선언과 구버전 호환성 확인
- [x] 독립 계획 리뷰 및 필요 수정
- [x] 사용자 계획 승인
- [x] 회귀 테스트 실패 확인
- [x] 설치, 템플릿, upgrade 경로 변경
- [x] 리팩터링 검토 및 문서 갱신
- [x] 전체 테스트, vet, 빌드와 생성 프로젝트 검증
- [x] 작업 기록 정리 및 원자적 커밋
