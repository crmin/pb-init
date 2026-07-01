# pb-init 구현 계획

## 목표와 현재 상태

- 목표: `go run github.com/crmin/pb-init [moduleName] [args...]` 방식으로 실행 가능한 PocketBase SDK 프로젝트 초기화 CLI를 `SPEC.md` 계약에 맞게 구현한다.
- 추가 목표: `go run github.com/crmin/pb-init` 방식으로 사용하는 사용자를 대상으로 하는 영어 `README.md`를 작성한다.
- 현재 상태:
  - 루트에는 `main.go` 등 실행 가능한 Go 소스가 없어 `go run . --help`가 `no Go files in /Users/crmin/workspace/crmin/pb-init`로 실패한다.
  - `go test ./...`는 `go: warning: "./..." matched no packages` 및 `no packages to test`로 실패한다.
  - `templates/` 아래 템플릿 파일은 존재하지만 CLI 파서, 프로젝트 준비 로직, 템플릿 렌더링 로직, embed 로직, 테스트, README가 없다.
  - `RTK.md`는 `AGENTS.md`에서 참조되지만 현재 작업트리에 존재하지 않는다.

## 현재 계약 근거

- 실행 형식: `SPEC.md:3`-`SPEC.md:6`
- `moduleName`이 있으면 마지막 경로 요소로 하위 디렉토리를 만들고 원래 module path로 `go mod init` 실행: `SPEC.md:12`-`SPEC.md:23`
- `moduleName`이 없을 때 현재 디렉토리의 Go module 판정 조건: `SPEC.md:25`-`SPEC.md:30`
- 기존 Go module에서 `go.sum` 또는 `*.go`가 있으면 `--force` 요구 및 고정 메시지 출력: `SPEC.md:31`-`SPEC.md:42`
- Go module이 아닌 현재 디렉토리에서 moduleName 누락 시 동적 module path 기반 예시 메시지 출력: `SPEC.md:43`-`SPEC.md:52`
- optional arguments: `--migration-dir`, `--pb-version`: `SPEC.md:56`-`SPEC.md:59`
- short flag bundle 및 help message 규칙: `SPEC.md:61`-`SPEC.md:92`
- PocketBase SDK 설치, 템플릿 렌더링, migration directory 생성, Docker/.gitignore 생성: `SPEC.md:94`-`SPEC.md:110`
- 템플릿 embed 요구: `SPEC.md:112`-`SPEC.md:114`
- 현재 템플릿 위치:
  - `templates/main.go.tmpl`
  - `templates/migration_init.go.tmpl`
  - `templates/Dockerfile.tmpl`
  - `templates/.dockerignore.tmpl`
  - `templates/.gitignore.tmpl`

## 계획 승인 시 함께 고정할 정적 메시지

### Help Message

`{command}`는 `debug.ReadBuildInfo()`에서 확인한 main module path를 사용한다.

```text
PocketBase project initializer

Usage:
  go run {command} [moduleName] [options] [flags]

Arguments:
  moduleName
    Optional Go module path. When provided, pb-init creates a directory named after the last path element and runs `go mod init <moduleName>` in it. When omitted, the current directory must already be a Go module.

Options:
  --migration-dir=<dirName>
    Directory for generated Go migration files that are compiled into the Go build. The value may be nested, for example `internal/migrations`, but it must be a child path relative to the PocketBase project module directory: absolute paths, current directory references (`.`), and parent directory references (`..`) are rejected. When `--jsvm` is used, this option still only controls Go migrations; JavaScript migrations remain in `pb_migrations` and JavaScript hooks remain in `pb_hooks`. Default: migrations
  --pb-version=<version>
    PocketBase SDK version passed to `go get github.com/pocketbase/pocketbase@<version>`. Default: latest

Flags:
  -h, --help
    Show this help message and exit.
  --force
    Initialize the current Go module even when go.sum or Go source files already exist.
  -d, --docker
    Generate Dockerfile and .dockerignore.
  -m, --auto-migration
    Enable PocketBase auto migration in generated code.
  -j, --jsvm
    Enable PocketBase JS hooks, JS migrations, and JS runtime support.
  --cgo-enabled
    Use CGO_ENABLED=1 in the generated Dockerfile when --docker is enabled.
  -r, --recommend
    Equivalent to --docker --auto-migration.
```

### Error And Warning Messages

명세에 고정된 메시지는 줄바꿈과 마크다운 문자를 그대로 사용한다.

출력 채널은 아래로 고정한다.

- `--help`로 요청된 help message는 stdout에 출력하고 exit code 0으로 종료한다.
- 에러 발생 시 CLI가 생성하는 오류 메시지와 help message는 stderr에 출력하고 exit code 1로 종료한다.
- 외부 명령 실패 시 전달하는 command output은 stderr에 출력하고 exit code 1로 종료한다.
- `--force`가 제공되어 기존 Go module 초기화를 계속 진행한다는 안내 메시지는 오류가 아니므로 stdout에 출력한다.

```text
This directory is already initialized as a Go module. To initialize the current directory as a PocketBase project, run the command again with the `--force` flag.
**Warning**: This may overwrite or damage your existing project.
```

```text
This directory is already initialized as a Go module. Since the `--force` flag was provided, PocketBase project initialization will proceed.
Warning: Existing project files may be overwritten or corrupted.
```

```text
This directory is not initialized as a Go module. To initialize the current directory as a PocketBase project, please provide a module name as an argument.

Example:
- `go run {command} myproject`
- `go run {command} github.com/username/myproject`
```

```text
Invalid flag: -{flag}

{help message}
```

`-{flag}`는 short flag bundle에서 실제로 입력된 잘못된 short flag 문자로 대체한다. 예를 들어 `-dmx`는 `Invalid flag: -x`를 출력한다.

```text
Invalid flag: -h
Cannot use -h in a short flag bundle.

{help message}
```

```text
Invalid flag: -r
Cannot use -r in a short flag bundle.

{help message}
```

명세에 직접 문구가 없는 파서 오류는 아래로 고정한다.

```text
Invalid flag: --unknown

{help message}
```

```text
Missing value for --migration-dir.

{help message}
```

```text
Missing value for --pb-version.

{help message}
```

```text
Invalid --pb-version: none is not allowed. Provide a PocketBase version or omit --pb-version to use latest.

{help message}
```

```text
Unexpected argument: {argument}

{help message}
```

```text
Invalid --migration-dir: absolute paths are not allowed. Use a child path relative to the PocketBase project module directory.

{help message}
```

```text
Invalid --migration-dir: current directory references (`.`) are not allowed. Use a child path relative to the PocketBase project module directory.

{help message}
```

```text
Invalid --migration-dir: parent directory references (`..`) are not allowed. Use a child path relative to the PocketBase project module directory.

{help message}
```

외부 명령 실패 중 `go get github.com/pocketbase/pocketbase@{pb-version}` 실패는 command output을 stderr에 그대로 출력하고 exit code 1로 종료한다. 이는 현재 `SPEC.md:96`의 stdout 출력 계약을 사용자 추가 요청에 맞춰 stderr 출력 계약으로 변경하는 항목이다. `go mod init` 실패는 현재 `SPEC.md`에 출력 채널 계약이 없으므로, 사용자 추가 요청에 따라 별도 prefix 없이 command output을 stderr에 그대로 출력하고 exit code 1로 종료하도록 새로 고정한다.

## 계획 승인 시 함께 승인되는 구현 결정

- `SPEC.md:37`의 `--flag`는 주변 문맥상 `--force` 오탈자로 해석한다.
- optional argument는 `--migration-dir=value`, `--migration-dir value`, `--pb-version=value`, `--pb-version value`를 모두 허용한다. 이는 일반적인 CLI 인식 범위이고 명세의 `--migration-dir={dirName}` 형식과 충돌하지 않는다.
- `--pb-version` 값으로 정확히 `none`이 전달되면 `Invalid --pb-version: none is not allowed. Provide a PocketBase version or omit --pb-version to use latest.`를 stderr에 출력하고 종료한다. `none` 외 값은 `go get github.com/pocketbase/pocketbase@{pb-version}`에 전달한다.
- `moduleName`은 첫 번째 positional argument로만 허용하고 추가 positional argument는 `Unexpected argument`로 실패시킨다.
- `moduleName`에는 명세에 없는 별도 선제 validation을 추가하지 않는다. directory name 산출과 `go mod init` 실행 중 OS 또는 Go toolchain이 실패하면 해당 command output을 stderr에 그대로 전달한다.
- `--cgo-enabled`는 `--docker`가 있을 때만 Dockerfile 렌더링에 영향을 주며, 단독 전달 시에는 오류 없이 무시한다.
- `--recommend`와 `-r`은 `--docker --auto-migration`과 동일하게 처리하고, `-r`은 명세대로 short flag bundle에 포함될 수 없다.
- `--migration-dir`는 PocketBase project module directory 기준 child relative path만 허용한다. 절대 경로는 "Invalid --migration-dir: absolute paths are not allowed. Use a child path relative to the PocketBase project module directory.", path component 중 `.`가 포함된 값은 "Invalid --migration-dir: current directory references (`.`) are not allowed. Use a child path relative to the PocketBase project module directory.", path component 중 `..`가 포함된 값은 "Invalid --migration-dir: parent directory references (`..`) are not allowed. Use a child path relative to the PocketBase project module directory."로 실패한다.
- migration directory가 `internal/migrations`처럼 중첩 경로일 수 있으므로, migration `init.go`의 package name은 directory base를 Go package identifier로 변환해서 사용한다. 이를 위해 `templates/migration_init.go.tmpl`에는 `{{.MigrationPackage}}` 변수를 추가하고, `SPEC.md`도 이 추가 template variable을 반영하도록 보완한다. 이는 중첩 directory 지원을 실제 Go code로 성립시키기 위한 명세 보완이다.
- `--jsvm`은 구현 시점의 `go get github.com/pocketbase/pocketbase@latest` 결과에 맞는 `jsvm.MustRegister(app, jsvm.Config{})` API로 실제 JS hooks, JS migrations, JS runtime을 활성화한다. 계획 조사 시점의 latest는 `v0.39.5`였지만 구현은 특정 버전에 pin하지 않는다. `--migration-dir`는 Go migration directory에만 사용하고 JS migrations는 PocketBase 기본값인 `pb_migrations`를 따른다.
- `--jsvm`은 generated project가 즉시 빌드될 수 있도록 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}`를 추가로 실행한다. 실제 smoke에서 `go get github.com/pocketbase/pocketbase@latest`만으로는 jsvm plugin의 transitive dependency go.sum 항목이 부족해 generated project build가 실패함을 확인했기 때문이다.
- `--jsvm`이 전달되면 PocketBase project module directory에 `pb_migrations`와 `pb_hooks` 빈 디렉토리를 생성한다. 이미 존재하면 유지한다.
- `--jsvm --docker` 조합에서는 generated Dockerfile의 final stage에 JS migration directory `pb_migrations`와 JS hooks directory `pb_hooks`가 포함되어야 한다. `pb_hook`이 아니라 PocketBase 기본 directory 이름인 `pb_hooks`를 사용한다.

## 변경 예정 파일과 내용

### 루트 파일

- `main.go`
  - `package main` 진입점 추가.
  - `//go:embed templates/*`로 `templates/` 하위 모든 템플릿 파일 embed.
  - `internal/initcli`의 실행 함수에 `os.Args[1:]`, stdout, stderr, 현재 작업 디렉토리, embed FS, command runner를 주입.
  - 반환 exit code로 `os.Exit` 호출.
- `README.md`
  - 영어로 작성.
  - `go run github.com/crmin/pb-init` 원격 실행 사용자를 기준으로 quick start, current directory initialization, moduleName initialization, flags/options, Docker generation, JSVM behavior, `--force` warning, generated files, examples, troubleshooting 작성.
- `SPEC.md`
  - `--migration-dir`는 target module 내부 상대 경로만 허용하고 절대 경로와 parent directory reference를 거부한다는 계약을 보완한다.
  - `--migration-dir` 값으로 current directory reference `.`를 거부한다는 계약을 보완한다.
  - `--pb-version` 값으로 `none`을 금지한다는 계약과 고정 오류 메시지를 보완한다.
  - 에러 발생 시 stderr 출력 계약을 보완한다.
  - `SPEC.md:96`의 `go get` 실패 시 stdout 출력 계약을 stderr 출력 계약으로 변경한다.
  - `go mod init` 실패 시 command output을 stderr로 전달하는 계약을 추가한다.
  - `templates/migration_init.go.tmpl`에 `{{.MigrationPackage}}` 변수가 추가되는 이유와 값을 보완한다.
  - `--jsvm` 전달 시 `pb_migrations`, `pb_hooks` 빈 디렉토리를 생성한다는 계약을 보완한다.
  - `--jsvm --docker` 조합에서 Dockerfile final stage에 `pb_migrations`, `pb_hooks`를 포함한다는 계약과 Dockerfile template variable을 보완한다.
  - 기존 동작 계약을 바꾸지 않고, 중첩 migration directory를 유효한 Go package로 생성하기 위한 template variable 설명을 추가한다.
- `go.mod`
  - 기본적으로 변경하지 않는다. 표준 라이브러리만 사용하도록 설계한다.

### 내부 구현

- `internal/initcli/cli.go`
  - `Config`, `Env`, `ParseArgs`, `Run` 정의.
  - help message와 고정 error/warning message 상수 정의.
  - long flag, optional argument, short flag bundle 파싱.
  - `--pb-version` 값이 `none`이면 고정 메시지로 stderr에 출력되도록 validation.
  - `--migration-dir`가 절대 경로이거나 path component에 `.` 또는 `..`를 포함하면 고정 메시지로 stderr에 출력되도록 validation.
  - `debug.ReadBuildInfo()` 기반 command module path 확인.
- `internal/initcli/project.go`
  - 현재 디렉토리 Go module 판정.
  - `moduleName` 기반 프로젝트 디렉토리 결정 및 `go mod init` 실행.
  - current directory 초기화 시 `--force` 필요 여부 판단.
  - `go get github.com/pocketbase/pocketbase@{pb-version}` 실행.
  - `--jsvm`일 때 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}` 추가 실행.
  - `go.mod`에서 module path 읽기.
- `internal/initcli/render.go`
  - embed FS에서 템플릿 로드 및 렌더링.
  - `main.go`, migration `init.go`, `.gitignore`, 선택적 `Dockerfile`, 선택적 `.dockerignore` 생성.
  - nested migration directory 생성.
  - `--jsvm`이 true이면 `pb_migrations`, `pb_hooks` 빈 디렉토리를 exist-ok로 생성.
  - `BinaryName`, `CgoEnabled`, `MigrationPackage`, `AutoMigration`, `JSVMImport`, `JSVMAssets` 값 계산.
- `internal/initcli/*_test.go`
  - 인자 파싱, module 판정, force 동작, command runner 호출, 템플릿 렌더링, 파일 생성 테스트 추가.
  - `--jsvm`일 때 jsvm plugin dependency를 추가 `go get`하는 테스트 추가.

### 템플릿

- `templates/main.go.tmpl`
  - `--jsvm` 활성화 시 import만 추가하지 않고 `jsvm.MustRegister(app, jsvm.Config{})` 호출 추가.
  - import formatting이 `gofmt` 후 유효하도록 템플릿 공백 조정.
  - `AutoMigration`은 명세대로 문자열 `"true"`/`"false"`를 렌더링해 생성된 코드에서는 bool literal이 되도록 유지.
- `templates/migration_init.go.tmpl`
  - `package {{.MigrationDir}}`에서 `package {{.MigrationPackage}}`로 변경.
  - 같은 commit에서 `SPEC.md`의 template variable 설명을 보완한다.
- `templates/Dockerfile.tmpl`
  - `{{.CgoEnabled}}` 값이 `"0"` 또는 `"1"`로 렌더링되는지 테스트로 고정한다. 필요 시 템플릿 자체는 유지한다.
  - `{{.JSVMAssets}}` 값이 true일 때 builder stage에서 `RUN mkdir -p pb_migrations pb_hooks`를 실행해 directory가 없더라도 final stage copy가 실패하지 않도록 한다.
  - `{{.JSVMAssets}}` 값이 true일 때 final stage에 아래 두 줄을 렌더링한다.

    ```Dockerfile
    COPY --from=builder /go/src/app/pb_migrations /pb_migrations
    COPY --from=builder /go/src/app/pb_hooks /pb_hooks
    ```

  - `{{.JSVMAssets}}` 값이 false일 때는 위 `RUN mkdir -p ...`와 `COPY` 두 줄을 렌더링하지 않는다.
- `templates/.dockerignore.tmpl`, `templates/.gitignore.tmpl`
  - `{{.BinaryName}}` 처리 테스트를 추가하고, 필요 시 trailing newline만 정리한다.

### 작업 문서

- `docs/plans/2026-07-01-pb-init-implementation.md`
  - 이 계획 문서.
  - 구현 중 checklist 상태를 계속 갱신.
- `docs/works/pb-init-implementation/learnings.md`
  - 조사 결과, 명세 근거, PocketBase API 확인 결과 기록.
- `docs/works/pb-init-implementation/decisions.md`
  - 현재 유효한 결정과 변경 이력 기록.
- `docs/works/pb-init-implementation/issues.md`
  - 막힘, 실패, 재시도 정책 기록.
- `docs/works/pb-init-implementation/problems.md`
  - 현재 문제 정의와 회귀 방지 수단 기록.

## TDD 계획

1. CLI 파싱 테스트를 먼저 추가하고 실패 확인.
   - `TestParseHelpIgnoresModuleName`
   - `TestParseShortFlagBundleExpandsDockerAutoMigrationJSVM`
   - `TestParseShortBundleRejectsHelp`
   - `TestParseShortBundleRejectsRecommend`
   - `TestParseShortBundleRejectsUnknownFlag`
   - `TestParseRecommendExpandsDockerAutoMigration`
   - `TestParseOptionsAcceptEqualsAndSeparateValues`
   - `TestParseUnknownLongFlag`
   - `TestParseRejectsUnexpectedArgument`
   - `TestParseRejectsNonePBVersionToStderr`
   - `TestParseRejectsAbsoluteMigrationDirToStderr`
   - `TestParseRejectsCurrentMigrationDirToStderr`
   - `TestParseRejectsParentMigrationDirToStderr`
   - `TestParseInvalidShortFlagUsesInputFlagCharacter`
   - `TestErrorsWriteToStderr`
2. 현재 디렉토리 및 module 준비 테스트를 추가하고 실패 확인.
   - `TestIsGoModuleRequiresModuleFirstLineAndGoVersion`
   - `TestIsGoModuleAcceptsGoPatchVersion`
   - `TestCurrentModuleWithoutGoSumOrGoFilesUsesCurrentDirectory`
   - `TestCurrentModuleWithGoSumRequiresForce`
   - `TestCurrentModuleWithGoFilesRequiresForce`
   - `TestCurrentModuleWithForcePrintsWarning`
   - `TestMissingModuleNameOutsideGoModuleUsesBuildModulePath`
   - `TestModuleNameCreatesLastPathDirectoryAndRunsGoModInit`
   - `TestGoGetFailureWritesCommandOutputToStderr`
3. 템플릿 및 파일 생성 테스트를 추가하고 실패 확인.
   - `TestRenderMainTemplateUsesModulePathMigrationDirJSVMAndAutoMigration`
   - `TestRenderMigrationInitUsesPackageNameFromNestedDir`
   - `TestRenderCreatesNestedMigrationDirectory`
   - `TestRenderDockerFilesUseCgoAndBinaryName`
   - `TestRenderCreatesJSVMAssetDirectoriesWhenJSVMEnabled`
   - `TestRenderSkipsJSVMAssetDirectoriesWhenJSVMDisabled`
   - `TestRenderDockerfileCopiesJSVMAssetDirectoriesWhenJSVMEnabled`
   - `TestRenderDockerfileOmitsJSVMAssetDirectoriesWhenJSVMDisabled`
   - `TestRenderBinaryNamePocketBaseIsOmitted`
   - `TestRenderAlwaysWritesGitignore`
   - `TestEmbeddedTemplatesIncludeAllRequiredFiles`
4. 최소 구현으로 테스트 통과.
5. `go test ./...` 유지 상태에서 필요한 refactoring 수행.
6. 실제 temp directory에서 `go run . github.com/crmin/pb-init-smoke --docker -mj --migration-dir=internal/migrations`를 실행하고 생성된 프로젝트에서 `go build ./...` 확인.

## 예상 동작

- `go run github.com/crmin/pb-init myproject`는 현재 디렉토리 아래 `myproject`를 만들고 해당 디렉토리에서 `go mod init myproject`, `go get github.com/pocketbase/pocketbase@latest`, 템플릿 파일 생성을 수행한다.
- `go run github.com/crmin/pb-init github.com/crmin/test-data`는 `test-data` 디렉토리를 만들고 `go mod init github.com/crmin/test-data`를 수행한다.
- `moduleName` 없이 실행하면 현재 디렉토리의 `go.mod`를 검사하고, 안전한 빈 Go module이면 현재 디렉토리를 초기화한다.
- 현재 Go module에 `go.sum` 또는 루트 `*.go`가 있으면 `--force` 없이는 고정 경고 메시지 후 종료한다.
- `--force`가 있으면 고정 경고 메시지를 출력한 뒤 계속 진행한다.
- `-dmj`는 `--docker --auto-migration --jsvm`과 동일하다.
- `-h`, `-r`은 단독 short flag로는 동작하지만 short flag bundle 안에서는 고정 오류 메시지로 실패한다.
- short flag bundle에서 잘못된 문자가 있으면 실제 입력 문자를 사용해 `Invalid flag: -{flag}` 오류를 stderr에 출력한다.
- `--pb-version=none`은 고정 오류 메시지를 stderr에 출력하고 종료한다.
- `--migration-dir`가 절대 경로이거나 `.` 또는 `..` path component를 포함하면 고정 오류 메시지를 stderr에 출력하고 종료한다.
- `--docker`가 있으면 `Dockerfile`과 `.dockerignore`를 생성하고, `--cgo-enabled`에 따라 `CGO_ENABLED=0` 또는 `CGO_ENABLED=1`을 렌더링한다.
- `--jsvm`이 있으면 PocketBase project module directory에 `pb_migrations`, `pb_hooks` 빈 디렉토리를 생성한다.
- `--jsvm`이 있으면 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}`를 추가 실행해 generated project의 jsvm import가 바로 빌드되도록 한다.
- `--docker --jsvm`이 있으면 generated Dockerfile final stage에 `/pb_migrations`, `/pb_hooks`가 포함된다. `--jsvm`이 없으면 해당 directory copy는 렌더링하지 않는다.
- `.gitignore`는 항상 생성한다.

## 부작용과 호환성 리스크

- CLI는 대상 프로젝트 디렉토리에 `main.go`, migration `init.go`, `.gitignore`, 선택적 Docker 파일을 쓴다. 기존 파일이 있으면 덮어쓸 수 있다.
- `go get github.com/pocketbase/pocketbase@latest`는 네트워크와 현재 시점의 PocketBase 최신 버전에 의존한다.
- `--jsvm`은 추가 `go get github.com/pocketbase/pocketbase/plugins/jsvm@{pb-version}` 명령을 실행하므로 네트워크 호출이 하나 더 발생한다.
- PocketBase 최신 API가 바뀌면 `--jsvm` 생성 코드가 실패할 수 있으므로 구현 완료 시 latest 기준 generated project build를 검증한다.
- `--migration-dir` 절대 경로, `.` path component, `..` path component는 target module 외부 또는 module root 자체를 migration directory로 쓰는 상황을 막기 위해 CLI validation 단계에서 거부한다.
- package name sanitization 때문에 directory base와 실제 package identifier가 다를 수 있다. Go에서는 정상적인 패턴이지만 README에 nested/custom migration directory 사용 시 generated package name이 자동 보정된다는 점을 언급한다.
- `--jsvm`은 `pb_migrations`, `pb_hooks` 빈 디렉토리를 생성하므로 초기화 결과에 빈 디렉토리가 생긴다. 이는 JS migrations/hooks 기본 위치를 명시하기 위한 의도된 side effect다.
- `--docker --jsvm`에서 Dockerfile은 `pb_migrations`, `pb_hooks`를 final image root로 copy한다. 빈 directory도 copy 가능하도록 builder stage에서 directory를 보장하지만, 사용자가 다른 runtime layout을 기대했다면 README로 기본 layout을 명확히 안내한다.

## 검증 계획

- 단위 테스트: `go test ./...`
- 빌드: `go build ./...`
- help smoke: `go run . --help`
- stderr smoke: invalid flag, `--pb-version=none`, invalid `--migration-dir` 실행 시 오류가 stderr로 출력되는지 확인
- temp directory 실제 생성 smoke:
  - 임시 디렉토리에서 `go run /Users/crmin/workspace/crmin/pb-init github.com/crmin/pb-init-smoke --docker -mj --migration-dir=internal/migrations`
  - 생성된 `pb-init-smoke` 디렉토리에서 `go build ./...`
  - 생성된 `main.go`, `internal/migrations/init.go`, `Dockerfile`, `.dockerignore`, `.gitignore` 확인
  - 생성된 `pb_migrations`, `pb_hooks` 디렉토리 확인
  - 생성된 `Dockerfile`에 `RUN mkdir -p pb_migrations pb_hooks`, `COPY --from=builder /go/src/app/pb_migrations /pb_migrations`, `COPY --from=builder /go/src/app/pb_hooks /pb_hooks`가 포함되는지 확인
- current directory mode smoke:
  - 임시 디렉토리에서 `go mod init example.com/current`
  - `go run /Users/crmin/workspace/crmin/pb-init --pb-version=latest`
  - 생성된 프로젝트에서 `go build ./...`
- `--force` failure/success는 unit test와 temp fixture 중심으로 검증한다.
- Docker image build는 Docker daemon availability에 따라 선택 검증으로 둔다. 기본 완료 조건은 생성 Dockerfile content와 generated project build다.

## Commit 계획

### Commit 1: CLI 파싱과 고정 메시지

- 변경/작성 파일:
  - `SPEC.md`
  - `main.go`
  - `internal/initcli/cli.go`
  - `internal/initcli/cli_test.go`
  - `docs/plans/2026-07-01-pb-init-implementation.md`
  - `docs/works/pb-init-implementation/learnings.md`
  - `docs/works/pb-init-implementation/decisions.md`
  - `docs/works/pb-init-implementation/issues.md`
  - `docs/works/pb-init-implementation/problems.md`
- 구체 내용:
  - `--migration-dir` 경로 제한, `--pb-version=none` 금지, stderr 출력 계약, `go get` 실패 출력 채널 변경, `go mod init` 실패 출력 채널 추가를 `SPEC.md`에 보완.
  - root main skeleton, embed wiring, parser, help/error/warning constants, stderr routing, `--pb-version` validation, `--migration-dir` validation, parser tests.
  - 계획 문서와 작업 문서의 초기 상태 포함.
- 예상 commit 시점:
  - parser tests와 `go test ./...` 통과 직후 즉시 commit.
- 예상 commit message:
  - `feat(cli): 인자 파싱과 고정 메시지 추가`

### Commit 2: Go module 준비와 PocketBase 설치

- 변경/작성 파일:
  - `internal/initcli/cli.go`
  - `internal/initcli/project.go`
  - `internal/initcli/project_test.go`
  - `docs/plans/2026-07-01-pb-init-implementation.md`
  - `docs/works/pb-init-implementation/learnings.md`
  - `docs/works/pb-init-implementation/decisions.md`
  - `docs/works/pb-init-implementation/issues.md`
  - `docs/works/pb-init-implementation/problems.md`
- 구체 내용:
  - 현재 디렉토리 Go module 판정, force guard, moduleName 디렉토리 생성, `go mod init`, `go get`, module path 읽기.
  - command runner fake 기반 테스트.
- 예상 commit 시점:
  - module/project tests와 `go test ./...` 통과 직후 즉시 commit.
- 예상 commit message:
  - `feat(init): Go 모듈 준비와 PocketBase 설치 처리 추가`

### Commit 3: 템플릿 렌더링과 프로젝트 파일 생성

- 변경/작성 파일:
  - `SPEC.md`
  - `internal/initcli/cli.go`
  - `internal/initcli/project.go`
  - `internal/initcli/render.go`
  - `internal/initcli/project_test.go`
  - `internal/initcli/render_test.go`
  - `templates/main.go.tmpl`
  - `templates/migration_init.go.tmpl`
  - `templates/Dockerfile.tmpl`
  - `templates/.dockerignore.tmpl`
  - `templates/.gitignore.tmpl`
  - `docs/plans/2026-07-01-pb-init-implementation.md`
  - `docs/works/pb-init-implementation/learnings.md`
  - `docs/works/pb-init-implementation/decisions.md`
  - `docs/works/pb-init-implementation/issues.md`
  - `docs/works/pb-init-implementation/problems.md`
- 구체 내용:
  - nested migration directory 지원을 위해 `SPEC.md`에 `MigrationPackage` template variable 설명 보완.
  - `--jsvm` generated build를 위해 jsvm plugin dependency 추가 `go get` 로직과 테스트 보완.
  - `--jsvm` 시 `pb_migrations`, `pb_hooks` 빈 디렉토리 생성 계약을 `SPEC.md`에 보완.
  - `--jsvm --docker` 조합에서 `pb_migrations`, `pb_hooks`를 final image에 포함하도록 `SPEC.md`와 `Dockerfile.tmpl` 보완.
  - 템플릿 렌더링, nested migration directory 생성, Docker/.dockerignore/.gitignore 생성.
  - `--jsvm` generated code compile 가능하도록 template 업데이트.
  - 렌더링 테스트 및 generated project build smoke.
- 예상 commit 시점:
  - render tests, `go test ./...`, `go build ./...`, generated project smoke build 통과 직후 즉시 commit.
- 예상 commit message:
  - `feat(init): 템플릿 기반 프로젝트 파일 생성 추가`

### Commit 4: README 작성과 최종 검증

- 변경/작성 파일:
  - `README.md`
  - `docs/plans/2026-07-01-pb-init-implementation.md`
  - `docs/works/pb-init-implementation/learnings.md`
  - `docs/works/pb-init-implementation/decisions.md`
  - `docs/works/pb-init-implementation/issues.md`
  - `docs/works/pb-init-implementation/problems.md`
- 구체 내용:
  - 영어 README 작성.
  - 최종 검증 결과와 체크리스트 반영.
- 예상 commit 시점:
  - README 작성 후 `go test ./...`, `go build ./...`, `go run . --help`, generated project smoke 통과 직후 즉시 commit.
- 예상 commit message:
  - `docs(readme): 원격 실행 사용 가이드 추가`

## TODO 체크리스트

- [x] `SPEC.md`, `AGENTS.md`, `planner`, `working-docs` 지침 확인
- [x] 현재 파일 구조와 현재 실행 실패 상태 확인
- [x] 계획 문서 초안 작성
- [x] subagent에게 `SPEC.md`와 계획 문서를 직접 읽게 하고 명세 커버리지 검토
- [x] 1차 subagent reject 결과 반영
- [x] 수정된 계획을 새 subagent로 재검토
- [x] 사용자 추가 변경 요청을 계획에 반영
- [x] 변경된 계획을 새 subagent로 재검토
- [x] 3차 subagent reject 결과 반영
- [x] 변경된 계획을 새 subagent로 재검토
- [x] 사용자 추가 변경 요청 2차 반영
- [x] 변경된 계획을 새 subagent로 재검토
- [ ] 사용자 승인 받기
- [x] Commit 1 구현 및 즉시 commit
- [x] Commit 2 구현 및 즉시 commit
- [x] Commit 3 구현 및 즉시 commit
- [ ] Commit 4 구현 및 즉시 commit
- [ ] 최종 상태 보고

## Subagent 검토 상태

- 1차 검토: REJECT.
  - 명세에 없는 `--migration-dir` 경로 제한 제거.
  - 명세에 없는 `moduleName` validation 제거.
  - `MigrationPackage` 추가는 `SPEC.md` 보완 예정 항목으로 명시.
  - PocketBase `v0.39.5`는 pin이 아니라 계획 시점 확인 정보로 정정.
- 2차 검토: APPROVE.
  - 필수 동작, 인자, flag, 고정 메시지, template 변수, embed/build 요구가 계획에 반영되어 있음을 확인.
  - commit 계획과 영어 README 계획도 요청 조건을 충족함을 확인.
  - 사소한 제안: `MigrationPackage` SPEC 보완 문구와 `--cgo-enabled` README 설명을 구현 시 명확히 작성.
- 3차 검토: REJECT.
  - `go get` 실패 출력 채널을 stderr로 바꾸는 것이 현재 `SPEC.md:96`의 stdout 계약 변경임을 계획에 명확히 쓰도록 지적.
  - `go mod init` 실패 출력 채널은 `SPEC.md`에 없는 새 계약임을 분리해 쓰도록 지적.
- 4차 검토: APPROVE.
  - `go get` 실패 출력 채널 변경, `go mod init` 실패 출력 채널 추가, `--migration-dir` 제한과 고정 메시지, stderr 출력, Dockerfile JSVM asset copy, help/error 문구, commit 계획, README 계획이 요청을 cover함을 확인.
  - 사소한 제안: 구현 단계에서 `SPEC.md:96`의 기존 stdout 문구를 남기지 않고 stderr 문구로 완전히 교체.
- 5차 검토: APPROVE.
  - `--pb-version=none` 금지와 고정 메시지, `--migration-dir` 절대 경로/current directory reference/parent directory reference 금지, `--jsvm` 시 `pb_migrations`와 `pb_hooks` 빈 디렉토리 생성, 기존 stderr/Dockerfile/README/commit 계획이 요청을 cover함을 확인.
  - 사소한 제안: 구현 단계에서 `SPEC.md:96`의 기존 stdout 문구를 stderr 계약으로 완전히 교체.
