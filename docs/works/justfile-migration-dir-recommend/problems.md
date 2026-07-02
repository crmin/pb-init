# Problems

## 문제 정의

- `just snapshot`은 현재 `migrations` 디렉토리만 정리한다.
- `--migration-dir`로 다른 경로를 지정하면 PocketBase migrate command는 해당 경로에 migration file을 생성하지만, justfile cleanup은 다른 디렉토리를 본다.
- `--recommend`/`-r`은 현재 `--just`를 포함하지 않아 recommended project에 `justfile`이 생성되지 않는다.

## 재현 절차

```sh
go run . example.com/custom-migrations --just --migration-dir=internal/migrations
cd custom-migrations
just snapshot -y
```

예상 문제:

- migration file은 `internal/migrations`에 생성된다.
- cleanup logic은 `migrations`를 확인하므로 `internal/migrations`의 오래된 snapshot file을 정리하지 않는다.

```sh
go run . example.com/recommended --recommend
```

예상 문제:

- 현재 `--recommend`은 `justfile`을 생성하지 않는다.

## 원인 후보

- `templates/justfile.tmpl`이 `{{.MigrationDir}}` 대신 `migrations`를 직접 사용한다.
- `applyRecommend`가 `cfg.Just`를 true로 설정하지 않는다.
- `SPEC.md`, `README.md`, help text가 이전 `--recommend` 계약을 설명한다.

## 회귀 방지 수단

- custom migration dir justfile 렌더링 테스트를 추가한다.
- recommend parser 테스트에서 `cfg.Just`까지 확인한다.
- help message, SPEC, README를 함께 갱신해 계약 불일치를 방지한다.

## 해결 내용 요약

- `templates/justfile.tmpl`에 `migration_dir="{{.MigrationDir}}"`를 렌더링하고 cleanup 경로 조회를 `"$migration_dir"` 기준으로 변경했다.
- `applyRecommend`가 `cfg.Just`도 true로 설정하도록 변경했다.
- help text, `SPEC.md`, `README.md`, 관련 작업 문서를 새 계약에 맞게 갱신했다.
- `go test ./...`, `go build ./...`, 임시 프로젝트 smoke를 통해 회귀 방지 테스트와 실제 생성 결과를 검증했다.
