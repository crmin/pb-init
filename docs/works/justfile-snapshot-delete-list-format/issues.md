# Issues

## Open

- 없음.

## Resolved

- 삭제 대상 파일 목록이 단순 줄 목록으로 출력됨.
  - 원인: `printf '%s\n' "${delete_files[@]}"` 사용.
  - 해결: `printf '    - %s\n' "${delete_files[@]}"`로 변경.
