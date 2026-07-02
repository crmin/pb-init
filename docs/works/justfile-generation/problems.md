# Problems

## 문제 정의

- 생성된 PocketBase project에는 반복 사용 명령을 모아둔 `justfile`이 없다.
- 현재 CLI에는 `--just` 옵션이 없어 사용자가 원할 때만 `justfile`을 생성할 수 없다.
- `.dockerignore`는 `justfile` 생성 여부를 알지 못해 Docker build context에서 `justfile`을 조건부 제외할 수 없다.
- `snapshot` command는 단순 command alias가 아니라 migration 생성 후 최신 snapshot 하나만 남기는 shell 로직이 필요하다.

## 재현 절차

현재 `--just` flag는 존재하지 않는다.

```sh
go run . example.com/app --just
```

예상되는 현재 결과:

```text
Invalid flag: --just
```

현재 `RenderProject`는 `justfile`을 생성하지 않는다.

## 원인 후보

- `Config`에 `Just` 설정값이 없다.
- `parseLongFlag`가 `--just`를 처리하지 않는다.
- `templates/justfile.tmpl`이 없다.
- `RenderProject`에 `cfg.Just` 기반 렌더링 분기가 없다.
- `.dockerignore.tmpl`에 justfile 조건부 항목을 렌더링할 template data가 없다.

## 회귀 방지 수단

- parser test로 `--just` flag 계약을 고정한다.
- render test로 `justfile` 생성/미생성과 `.dockerignore` 조건 렌더링을 고정한다.
- embed test로 신규 template이 binary에 포함되는지 확인한다.
- 가능한 환경에서는 `just --list`와 `just --dry-run` 기반 smoke test를 skip 가능한 테스트로 추가한다.
- subagent 검증 결과를 plan과 works 문서에 기록해 justfile shell 로직의 의사결정을 추적 가능하게 한다.

## 해결 내용 요약

- `--just` parser/help를 추가했다.
- `templates/justfile.tmpl`을 추가하고 `RenderProject`가 `cfg.Just`일 때 module root `justfile`을 렌더링하도록 변경했다.
- `.dockerignore.tmpl`에 `Justfile` 조건을 추가해 `--docker --just` 조합에서만 `justfile` 항목을 렌더링하도록 변경했다.
- `SPEC.md`와 `README.md`에 `--just`와 generated just command 계약을 반영했다.
