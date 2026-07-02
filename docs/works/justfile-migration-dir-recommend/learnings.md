# Learnings

## TL;DR

- 현재 `templates/justfile.tmpl`은 `snapshot` 정리 대상 경로로 `migrations`를 hardcoding한다.
- generated `main.go`는 `{{.MigrationDir}}` 값을 `migratecmd.Config.Dir`로 사용하므로 custom `--migration-dir`에서는 실제 migration 생성 위치와 justfile 정리 위치가 달라진다.
- 현재 `applyRecommend`는 `Docker`와 `AutoMigration`만 활성화하고 `Just`는 활성화하지 않는다.
- baseline `go test ./...`는 계획 작성 시점에 통과한다.
- 실패 테스트 추가 후 `--recommend`의 `Just=false`, help 문구, hardcoded `migrations` 경로가 예상대로 실패했다.
- 구현 후 `templates/justfile.tmpl`은 `migration_dir="{{.MigrationDir}}"`를 렌더링하고, `find "$migration_dir"`로 cleanup 대상을 조회한다.
- 최종 검증에서 `go test ./...`, `go build ./...`, `--recommend --migration-dir=internal/migrations` 임시 프로젝트 생성 smoke, `.dockerignore`의 `justfile` 항목, generated `justfile`의 `migration_dir="internal/migrations"`, `just snapshot` dry-run syntax를 확인했다.

## 근거

- 변경 전 `templates/justfile.tmpl`의 `snapshot` recipe는 `[[ ! -d migrations ]]`, `find migrations ...`를 사용했다.
- `templates/main.go.tmpl`은 migration import와 `migratecmd.Config.Dir`에 `{{.MigrationDir}}`를 사용한다.
- `internal/initcli/render.go`의 `renderData`에는 이미 `MigrationDir` field가 있으므로 justfile template에서 곧바로 사용할 수 있다.
- `internal/initcli/cli.go`의 `applyRecommend`는 현재 `cfg.Docker = true`, `cfg.AutoMigration = true`만 수행한다.
- `applyRecommend`에 `cfg.Just = true`를 추가하면 `--recommend`와 `-r`이 같은 경로로 `Just`를 활성화한다.
- `just --dry-run snapshot -- -y`는 dry-run 특성상 recipe body를 출력하지만 실제 `./pocketbase`나 삭제 명령은 실행하지 않아 syntax smoke에 적합하다.

## 재사용 키워드

- justfile
- migration-dir
- recommend
- applyRecommend
- migratecmd.Config.Dir
- snapshot cleanup
