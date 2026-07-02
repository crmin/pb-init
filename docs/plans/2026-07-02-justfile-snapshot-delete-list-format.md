# justfile snapshot 삭제 목록 포맷 개선 계획

## 목표와 현재 동작

- 목표: `just snapshot`에서 삭제 대상 파일 목록을 읽기 쉬운 indented bullet list로 출력한다.
- 현재 동작: `printf '%s\n' "${delete_files[@]}"`로 파일 경로만 줄 단위 출력한다.

## 관련 파일

- `templates/justfile.tmpl`: 삭제 대상 목록 출력 형식 변경.
- `internal/initcli/render_test.go`: generated justfile의 출력 형식 테스트.
- `SPEC.md`, `README.md`: 사용자 출력 계약 반영.
- `docs/works/justfile-snapshot-delete-list-format/`: 작업 기록.

## 변경 사항

- 삭제 대상 목록 출력:

```bash
printf 'The following files will be deleted:\n'
printf '    - %s\n' "${delete_files[@]}"
```

## 검증

- `go test ./...`
- `go build ./...`

## TODO 체크리스트

- [x] 현재 템플릿과 테스트 확인.
- [x] 삭제 목록 bullet 출력 테스트 추가.
- [x] `templates/justfile.tmpl` 출력 형식 변경.
- [x] SPEC/README/work docs 갱신.
- [x] 테스트와 빌드 실행.
- [x] commit 생성.
