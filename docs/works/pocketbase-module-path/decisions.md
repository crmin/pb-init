# 결정

## Current (Active)

- PocketBase 공식 Go 예제와 버전별 `go.mod`의 module 선언을 경로 판단의 근거로 사용한다.
- SDK, JSVM, Go import와 `just upgrade`는 한 번 결정한 모듈 경로를 사용한다. SDK와 JSVM에는 같은 확정 버전을 전달한다.
- 기본 `latest`는 새 경로에서 조회한다. `patch`와 `upgrade`는 기존 `go.mod`가 요구하는 경로를 우선한다.

## Change Log

### 2026-11-11

- Changed: 기본 `latest`는 새 경로에서 조회하고, 조회 결과의 module 경로와 버전을 SDK와 JSVM에 함께 전달하는 계획을 작성했다.
- Reason: v0.39.5를 새 경로에서 설치하면 Go가 경로 불일치로 종료한다. v0.41.0-rc.1은 이미 새 경로를 선언한다.
- Changed: 승인 후 조회 결과를 설치와 템플릿에 연결했다. 버전이나 리비전을 새 경로에서 찾지 못했을 때 GitHub 경로에서 다시 조회한다.
- Reason: 구버전과 새 릴리스의 import를 일치시키고, 프록시 장애가 발생하면 원래 오류를 전달하기 위함이다.
- Changed: 조회를 `-mod=readonly`로 실행하고 Go 진단에서 프록시 상태와 응답 본문을 구분한다.
- Reason: 자동 vendor 모드에서 메타데이터를 조회하고, 파일 및 HTTP 프록시의 버전 부재에서만 GitHub 경로를 재조회하기 위함이다.
