# Issues

## Open

- `RTK.md`가 `AGENTS.md`에서 참조되지만 현재 작업트리에서 발견되지 않는다.
  - 영향: 추가 프로젝트 지침이 있을 가능성은 있지만 로컬에서 확인할 수 없다.
  - 현재 대응: 확인 가능한 `AGENTS.md`, `SPEC.md`, skill 지침을 기준으로 계획을 작성한다.

- `snapshot`은 generated `justfile`의 `migrations` 디렉토리만 정리한다.
  - 영향: `--migration-dir=internal/migrations`처럼 custom migration directory를 쓰는 프로젝트에서는 `just snapshot`이 생성된 custom directory 파일을 정리하지 않는다.
  - 현재 대응: 사용자 요구와 승인된 계획이 `migrations` 디렉토리를 명시했으므로 이번 작업에서는 범위를 확장하지 않는다.

## Resolved

- subagent 1차 검증에서 `snapshot` 빈 배열 확장 문제가 발견됐다.
  - 원인: bash 3.2와 `set -u` 조합에서 빈 `migrate_args` 배열을 `"${migrate_args[@]}"`로 확장할 때 실패할 수 있음.
  - 해결: `migrate_args` 길이가 0이면 인자 없는 command를 별도 분기로 실행하도록 justfile 후보를 수정.

- subagent 1차 검증에서 삭제 범위가 최신 snapshot 하나만 보존한다는 목표와 다를 수 있음이 지적됐다.
  - 원인: 기존 후보는 timestamp 형식 파일과 `init.go`만 삭제 대상으로 잡고 non-numeric `.go`는 남겼음.
  - 해결: numeric timestamp prefix가 가장 큰 `.go` 파일 하나를 제외한 `migrations/*.go` 전체를 삭제 대상으로 잡도록 수정.

- subagent 1차 검증에서 `-y`가 PocketBase 인자로 필요할 때 전달할 방법이 없다는 문제가 지적됐다.
  - 원인: 기존 후보는 모든 위치의 `-y`를 recipe 전용 flag로 제거했음.
  - 해결: `--` sentinel 뒤의 인자는 그대로 PocketBase migrate command에 전달하도록 수정.

- subagent 2차 검증에서 수정 후보가 승인 가능하다는 결과를 받았다.
  - 확인: `[private] default` 숨김, 공백 포함 인자 보존, bash 3.2 빈 배열 처리, `--` sentinel 처리, 최신 `.go` 하나 보존, prompt 반복, `upgrade` prefix 분기, command echo 숨김이 확인됨.

- prompt 문구 `The following files will be deleted. Continue? (Y/n): `는 관례상 빈 Enter를 yes로 암시할 수 있다.
  - 원인: 사용자 요구 prompt 문자열이 `(Y/n)`이지만 입력 규칙은 `y`, `n`, 다른 값 재프롬프트로 정의됨.
  - 해결: 사용자 요구와 승인된 계획에 맞춰 빈 입력도 재프롬프트하는 것으로 문서화하고 구현.
