# Problems

## 문제 정의

- `SPEC.md`는 원격 모듈 실행형 PocketBase 프로젝트 초기화 CLI 계약을 정의하지만, 현재 저장소에는 실행 가능한 Go source package가 없다.
- 현재 `templates/`는 존재하지만 이를 embed하고 렌더링하는 CLI 구현이 없다.

## 재현 절차

```sh
go run . --help
```

현재 결과:

```text
no Go files in /Users/crmin/workspace/crmin/pb-init
```

```sh
go test ./...
```

현재 결과:

```text
go: warning: "./..." matched no packages
no packages to test
```

## 원인 후보

- root package에 `main.go`가 없다.
- CLI 인자 파서와 project initialization orchestration 구현이 없다.
- 템플릿 embed 및 렌더링 구현이 없다.

## 회귀 방지 수단

- parser, module preparation, rendering 단위 테스트를 추가한다.
- 실제 temp directory에서 `go run` 기반 smoke test와 generated project `go build ./...`를 수행한다.
