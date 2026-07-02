# Learnings

## TL;DR

- 기존 `just snapshot` 삭제 목록은 파일 경로만 줄 단위로 출력해 목록 구조가 덜 명확했다.
- `printf '    - %s\n' "${delete_files[@]}"`를 사용하면 bash 배열을 유지하면서 요청한 bullet 형식으로 출력할 수 있다.

## 근거

- 사용자 요청 출력:
  - 기존: `migrations/file.go`
  - 기대: `    - migrations/file.go`
- fake `go` command를 사용한 `just snapshot` smoke에서 삭제 대상 목록이 `    - {file}` 형식으로 출력되고, 줄 시작이 `migrations/`인 기존 형식이 남지 않음을 확인했다.

## 재사용 키워드

- justfile
- snapshot
- delete_files
- bullet list
