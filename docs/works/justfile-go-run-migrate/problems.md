# Problems

## 문제 정의

- 생성 직후 프로젝트에서 `just migrate`를 실행하면 `./pocketbase` 파일이 없어 실패한다.

## 재현 절차

```sh
just migrate
```

관찰된 오류:

```text
./pocketbase: No such file or directory
```

## 원인

- `templates/justfile.tmpl`이 `migrate`와 `snapshot`에서 `./pocketbase migrate collections`를 호출한다.
- pb-init은 `pocketbase` binary를 생성하지 않으므로 해당 파일은 존재하지 않는다.

## 해결 내용 요약

- `migrate`와 `snapshot` recipe를 `go run . migrate collections` 기반으로 변경한다.
- generated justfile 테스트로 `./pocketbase` 의존이 제거됐음을 고정한다.
- 실제 생성 프로젝트에서 `just migrate` smoke를 실행해 `./pocketbase` 부재 오류가 재발하지 않음을 확인했다.
