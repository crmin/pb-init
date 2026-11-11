# 조사 결과

## 2026-11-11

- 공식 Go 예제는 `ganigeorgiev.com/pocketbase/pocketbase`를 import한다. [PocketBase Go 문서](https://pocketbase.io/docs/go-overview/)
- Go 모듈 프록시에서 신규 경로의 v0.39.5와 v0.40.6은 `github.com/pocketbase/pocketbase`를 선언한다. v0.41.0-rc.1, v0.41.0, v0.42.0은 신규 경로를 선언한다. [v0.40.6 go.mod](https://proxy.golang.org/ganigeorgiev.com/pocketbase/pocketbase/@v/v0.40.6.mod), [v0.41.0 go.mod](https://proxy.golang.org/ganigeorgiev.com/pocketbase/pocketbase/@v/v0.41.0.mod)
- 조회 시점의 `latest`는 새 경로에서 v0.42.0, GitHub 경로에서 v0.40.6으로 해석됐다.
- 임시 Go 1.27 모듈에서 `go list -m -json ganigeorgiev.com/pocketbase/pocketbase@v0.40`이 v0.40.6과 그 `GoMod` 경로를 반환했다. JSON의 `Path`는 조회에 사용한 신규 경로이므로 실제 module 선언은 `GoMod` 파일에서 읽어야 한다.
- SDK와 JSVM을 같은 버전으로 설치하고 `go mod tidy`와 `go build ./...`를 실행하는 검증이 v0.39.5와 v0.42.0의 임시 프로젝트에서 모두 통과했다.
- v0.41.0 배포 소스의 CHANGELOG는 import 교체 전후에 같은 문자열을 표시한다. Go 소스에는 공식 Go 예제와 `go.mod`가 사용하는 전체 모듈 경로를 적용한다.

## 구현 검증

- Go 1.27.3, darwin/arm64, just 1.46.0에서 `go test -count=1 ./...`, `go vet ./...`, `go build ./...`를 통과했다.
- 기본 설치와 `--jsvm --recommend`는 v0.42.0을 새 경로로 설치했다. v0.39.5와 v0.40 쿼리는 GitHub 경로를, v0.41.0-rc.1과 v0.41.0은 새 경로를 사용했다. 생성 프로젝트 6개의 `go build ./...`가 모두 통과했다.
- 기존 v0.39.5 프로젝트에서 `--force --pb-version=patch`로 v0.39.11을 설치하고 빌드했다. 별도로 작성한 마이그레이션 파일 내용도 비교했다.
- 생성한 새 경로의 v0.41.0 프로젝트에서 `just upgrade 0.41.1` 실행 후 모듈 버전과 빌드를 확인했다.
- README 변경 문단은 Humanize Korean light 경로에서 검증을 통과했다.
