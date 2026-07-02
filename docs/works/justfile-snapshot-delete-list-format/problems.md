# Problems

## 문제 정의

- `just snapshot` 삭제 확인 전 출력되는 파일 목록이 prompt와 시각적으로 잘 구분되지 않는다.

## 재현 절차

```sh
just snapshot
```

기존 출력:

```text
The following files will be deleted:
migrations/1782973428_collections_snapshot.go
migrations/init.go
```

## 원인

- generated justfile이 삭제 대상 배열을 `printf '%s\n'`로 직접 출력한다.

## 해결 내용 요약

- 삭제 대상 배열을 `printf '    - %s\n'`로 출력하도록 변경했다.
- generated justfile 테스트로 새 출력 형식을 고정했다.
- fake `go` command 기반 smoke로 실제 `just snapshot` 출력 형식을 확인했다.
