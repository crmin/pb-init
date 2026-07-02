# Decisions

## Current (Active)

- 계획 우선 진행 - `planner` 지침에 따라 production code는 사용자 승인 전 수정하지 않는다.
- 작업 문서 유지 - `working-docs` 지침에 따라 `docs/plans/`와 `docs/works/justfile-migration-dir-recommend/`에 기록한다.
- justfile template 경로 유지 - 사용자 요구대로 `templates/justfile.tmpl`를 유지하고 그 안에서 `{{.MigrationDir}}`를 사용한다.
- snapshot 정리 대상은 configured migration dir - generated app이 실제 migration을 생성하는 `--migration-dir` 경로와 cleanup 경로를 일치시킨다.
- `--recommend`은 `--docker --auto-migration --just`와 동일 - 사용자 요구에 따라 `applyRecommend`가 `cfg.Just`도 활성화해야 한다.
- 기존 justfile-generation 문서도 최신화 - 이전 open issue와 결정을 새 요구사항에 맞춰 변경 이력으로 남긴다.

## Change Log

### 2026-07-02

- Changed: custom `--migration-dir`와 `--recommend`의 `--just` 포함을 새 계획 범위로 정의.
- Reason: 사용자가 기존 justfile 생성 동작의 정리 대상과 recommend 확장을 명시적으로 요청함.

- Changed: `templates/justfile.tmpl`에서 `{{.MigrationDir}}`를 사용하기로 결정.
- Reason: render data에 이미 `MigrationDir`가 있고, generated app의 migrate command 설정과 just snapshot cleanup 경로를 일치시킬 수 있기 때문.

- Changed: `applyRecommend`가 `Docker`, `AutoMigration`, `Just`를 모두 활성화하도록 구현.
- Reason: `--recommend`와 `-r` 모두 `--just`를 포함해야 한다는 사용자 요구를 반영하기 위함.

- Changed: `templates/justfile.tmpl`의 snapshot cleanup 경로를 `migration_dir="{{.MigrationDir}}"` 기반으로 변경.
- Reason: generated app의 migration 생성 경로와 snapshot cleanup 경로를 일치시키기 위함.

- Changed: `main` branch에서 직접 구현하고 기능 커밋을 생성.
- Reason: 사용자가 별도 브랜치 대신 main branch 작업을 명시했기 때문.
