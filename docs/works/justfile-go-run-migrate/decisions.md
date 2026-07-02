# Decisions

## Current (Active)

- `just migrate`는 `go run . migrate collections`를 사용 - 생성 직후 prebuilt `./pocketbase` binary가 없어도 동작해야 하기 때문.
- `just snapshot`도 `go run . migrate collections`를 사용 - migrate command와 동일한 실행 기준을 유지하기 위함.
- configured migration directory cleanup은 유지 - 이전 요구사항의 `{{.MigrationDir}}` 기반 정리 동작과 충돌하지 않음.

## Change Log

### 2026-07-02

- Changed: generated justfile의 migrate/snapshot command를 `./pocketbase`에서 `go run .`로 변경.
- Reason: 생성 직후 프로젝트에는 `./pocketbase` binary가 없어서 `just migrate`가 실패하기 때문.
