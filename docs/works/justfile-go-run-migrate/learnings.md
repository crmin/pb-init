# Learnings

## TL;DR

- generated project에는 초기 상태에서 `./pocketbase` binary가 없다.
- `just migrate`가 `./pocketbase migrate collections`를 호출하면 생성 직후 `No such file or directory`로 실패한다.
- `go run . migrate collections`는 생성 프로젝트 root에서 실행하면 `--migration-dir` 경로를 프로젝트 아래 경로로 해석한다.

## 근거

- 사용자 재현 로그: `/var/folders/.../migrate: line 19: ./pocketbase: No such file or directory`.
- 임시 프로젝트 smoke에서 `go run . migrate collections` 실행 시 `pb_data`와 configured migration directory가 생성 프로젝트 root 아래에 생성됨을 확인했다.
- generated `justfile`의 shebang recipe 파일 경로는 temp path로 표시될 수 있지만, command는 just 실행 working directory에서 실행된다.
- 최종 smoke에서 생성 프로젝트의 `justfile`에 `go run . migrate collections "$@"`가 포함되고 `./pocketbase migrate collections`가 없음을 확인했다.
- `printf 'n\n' | just migrate`를 실제 실행해 `./pocketbase: No such file or directory` 오류가 재발하지 않음을 확인했다.

## 재사용 키워드

- justfile
- migrate
- snapshot
- go run
- pocketbase binary
