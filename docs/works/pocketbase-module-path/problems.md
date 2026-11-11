# 문제

## 2026-11-11

- 변경 전 pb-init은 SDK와 JSVM 설치, 생성되는 `main.go`와 `justfile`에 GitHub 경로를 사용했다. 기본 설치가 GitHub의 마지막 버전인 v0.40.6에 머물렀다.
- 임시 모듈에서 `go get ganigeorgiev.com/pocketbase/pocketbase@v0.39.5`를 실행하면 해당 릴리스의 module 선언이 GitHub 경로이므로 설치에 실패한다.
- 이전 모듈을 요구하는 프로젝트에서 `@patch`를 새 경로로 조회하면 현재 버전을 찾지 못한다. 기존 요구 경로에서 조회하도록 수정했다.
- 이전 경로를 import한 마이그레이션은 이전 모듈의 레지스트리에 등록된다. `--force`로 새 앱을 생성할 때 함께 전환할 파일을 README에 안내했다.
- 설치와 생성 코드에 동일한 경로를 전달한다. 회귀 테스트와 임시 프로젝트 빌드로 실제 릴리스의 module 선언과 일치하는지 확인했다.
