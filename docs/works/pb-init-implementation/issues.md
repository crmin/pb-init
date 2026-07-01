# Issues

## Open

- `RTK.md`가 `AGENTS.md`에서 참조되지만 현재 작업트리에 존재하지 않는다.
  - 영향: 추가 지침이 있을 가능성은 있지만 로컬에서 확인할 수 없다.
  - 현재 대응: 확인 가능한 `AGENTS.md`, `SPEC.md`, skill 지침을 기준으로 계획을 작성한다.

## Resolved

- 1차 subagent 검토가 REJECT를 반환했다.
  - 원인: 명세에 없는 `--migration-dir` 경로 제한, `moduleName` validation, `MigrationPackage`의 SPEC 불일치, `v0.39.5` pin처럼 보이는 문구.
  - 해결: 계획에서 명세 밖 제한을 제거하고, `MigrationPackage`는 `SPEC.md` 보완 예정 항목으로 명시했으며, PocketBase version은 latest 검증 기준으로 정정했다.
- 2차 subagent 검토가 APPROVE를 반환했다.
  - 확인: 수정된 계획은 `SPEC.md` 필수 동작과 commit 계획, 영어 README 요구를 cover한다.
  - 후속: 구현 시 `MigrationPackage` SPEC 보완 문구와 `--cgo-enabled` README 설명을 명확히 작성한다.
- 3차 subagent 검토가 REJECT를 반환했다.
  - 원인: `go get` 실패 출력 채널 변경이 기존 `SPEC.md` stdout 계약 변경임을 숨기는 표현이 있었고, `go mod init` 실패 stderr 출력은 명세에 없는 새 계약임을 분리하지 않았다.
  - 해결: 계획에서 `go get` 실패 출력은 `SPEC.md:96` 변경 항목으로, `go mod init` 실패 출력은 새 stderr 계약 추가 항목으로 명시했다.
- 4차 subagent 검토가 APPROVE를 반환했다.
  - 확인: 사용자 추가 요청과 명세 변경 계획, commit 계획, README 계획이 모두 cover되었다.
  - 후속: 구현 단계에서 `SPEC.md:96`의 기존 stdout 문구를 stderr 문구로 완전히 교체한다.
- 5차 subagent 검토가 APPROVE를 반환했다.
  - 확인: `--pb-version=none` 금지, `--migration-dir`의 `.`, `..`, absolute path 금지, `--jsvm` 디렉토리 생성 계획, 기존 stderr/Dockerfile/README/commit 계획이 모두 cover되었다.
  - 후속: 구현 단계에서 `SPEC.md:96`의 기존 stdout 문구를 stderr 문구로 완전히 교체한다.
- Commit 3 smoke에서 `--jsvm` generated project build가 jsvm transitive dependency go.sum 항목 누락으로 실패했다.
  - 원인: `go get github.com/pocketbase/pocketbase@latest`만으로는 `github.com/pocketbase/pocketbase/plugins/jsvm` import에 필요한 optional dependency go.sum 항목이 채워지지 않음.
  - 해결: `--jsvm`일 때 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}`를 추가 실행하도록 구현.
- Commit 3 smoke에서 `go run /Users/crmin/workspace/crmin/pb-init ...`를 임시 디렉토리에서 직접 실행하면 Go tool이 현재 디렉토리의 module을 찾지 못해 실패했다.
  - 원인: 로컬 absolute source path 방식은 임시 디렉토리의 module context 없이 실행되어 Go tool이 source module을 해석하지 못함.
  - 해결: 같은 실패를 반복하지 않고 로컬 binary를 빌드한 뒤 임시 디렉토리에서 실행해 생성 결과와 generated project build를 검증했다.
