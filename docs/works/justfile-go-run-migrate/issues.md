# Issues

## Open

- 없음.

## Resolved

- `just migrate`가 `./pocketbase` 부재로 실패.
  - 원인: generated justfile이 prebuilt binary를 전제로 함.
  - 해결: `go run . migrate collections`를 호출하도록 변경.
