이 문서는 remote module run으로 pocketbase sdk project initialize를 수행하기 위한 도구의 명세와 계약 사항을 정의함.

다음과 같은 형식으로 실행 할 수 있어야 함:
```
go run github.com/crmin/pb-init [moduleName] [args...]
```

# Arguments

## Positional Arguments

`moduleName`이 optional argument나 flag에 앞서 전달 될 수 있음. 만약 이 값이 전달된다면 현재 디렉토리 아래에 `moduleName` 경로의 가장 마지막 부분으로 디렉토리를 생성하고, 해당 디렉토리를 go module로 초기화 함.
예를 들어 다음과 같은 명령이 실행되었다면:
```
go run github.com/crmin/pb-init myproject
```
현재 디렉토리 아래에 `myproject`라는 디렉토리가 생성되고, 해당 디렉토리에서 `go mod init myproject`가 실행됨.

또는 다음과 같은 명령이 실행되었다면:
```
go run github.com/crmin/pb-init 'github.com/crmin/test-data'
```
현재 디렉토리 아래에 `test-data`라는 디렉토리가 생성되고, 해당 디렉토리에서 `go mod init github.com/crmin/test-data`가 실행됨.

만약 `moduleName`이 전달되지 않았다면, 현재 디렉토리 상황에 따라 동작이 결정되어야 함.
다음 과정을 통해 go module인지를 확인 할 수 있음:
- 현재 디렉토리에 `go.mod` 파일이 존재함
- `go.mod` 파일을 읽었을 때 첫번째 줄이 `module`로 시작함
- `go.mod` 파일에서 `^go \d+\.\d+(?:\.\d+)?$` 형식의 줄이 존재함 (예를 들어 다음과 같은 값 모두를 cover해야 함: `go 1.20`, `go 1.20.3`)

1. 현재 디렉토리가 go module인 경우
    - `go.sum` 파일이 존재하는 경우 또는 `*.go` 파일이 존재하는 경우 `--force` flag를 요구함. 만약 `--force` flag가 전달되지 않았다면 다음 에러 메시지를 출력하고 종료
        ```
        This directory is already initialized as a Go module. To initialize the current directory as a PocketBase project, run the command again with the `--force` flag.
        **Warning**: This may overwrite or damage your existing project.
        ```
        - `--flag`가 전달되었을 때 `go.sum` 파일이 존재하거나 `*.go` 파일이 존재하는 경우 다음 메시지를 출력하고 다음 작업을 이어서 진행 (`동작` section 참고)
        ```
        This directory is already initialized as a Go module. Since the `--force` flag was provided, PocketBase project initialization will proceed.
        Warning: Existing project files may be overwritten or corrupted.
        ```
    - `go.sum` 파일과 `*.go` 파일이 존재하지 않는 경우 현재 디렉토리를 go module 경로로 사용. 이미 초기화되어있으므로 별도 작업 수행할 필요 없이 다음 단계를 수행 할 수 있음. (`동작` section 참고)
2. 현재 디렉토리가 go module이 아닌 경우
    - go module 경로 정보가 없으므로 사용자에게 요청받아야 함. 다음 에러 메시지를 출력하고 종료.
        ```
        This directory is not initialized as a Go module. To initialize the current directory as a PocketBase project, please provide a module name as an argument.

        Example:
        - `go run github.com/crmin/pb-init myproject`
        - `go run github.com/crmin/pb-init github.com/username/myproject`
        ```
        - 이 때, `github.com/crmin/pb-init`은 고정 문자열이 아닌 `debug.ReadBuildInfo()`를 통해 동적으로 가져온 값으로 설정되어야 함 -- 프로젝트 module 경로가 추후 변경 될 수 있음을 고려해야 함

이렇게 생성/설정된 디렉토리를 프로젝트 모듈 디렉토리라고 하자.

## Optional Arguments

- `--migration-dir={dirName}`: migration 파일이 위치할 디렉토리 경로를 지정합니다. 이 값은 pocketbase project module directory를 기준으로 하는 하위 상대 경로여야 하며, 절대 경로, current directory reference(`.`), parent directory reference(`..`)를 포함할 수 없습니다. 주의: `--jsvm`과 함께 실행되는 경우에도 이 경로가 js migration directory에 영향을 주지 않습니다. 이 디렉토리에서는 go migration file이 관리되고 go build 결과에 포함됩니다. js migration 파일은 이 값과 무관하게 `pb_migrations` 경로에서 관리됩니다. default=`migrations`
- `--pb-version`: pocketbase sdk version을 지정합니다. default=`latest`. `none` 값은 사용할 수 없으며 다음 에러 메시지를 stderr로 출력하고 종료합니다.
    ```
    Invalid --pb-version: none is not allowed. Provide a PocketBase version or omit --pb-version to use latest.

    {help message}
    ```

## Flags

short flag가 존재하는 경우 괄호 안에 병기함. 괄호 안에 없는 flag는 사용빈도가 낮거나 위험성을 고려해서 short flag를 제공하지 않음.
`-h`와 `-r`를 제외한 short flag는 이어서 전달 될 수 있음. 예를 들어 `-dmj`는 `-d -m -j`와 동일하게 동작해야 함. 순서는 중요하지 않음. 만약 묶여서 전달된 short flag 중 하나라도 잘못된 값이 존재한다면 다음 에러 메시지를 출력하고 종료:
```
Invalid flag: -x

{help message}
```
`-x`는 실제 입력된 invalid short flag 문자로 대체되어야 함. 예를 들어 `-dmz`가 전달되면 `Invalid flag: -z`를 출력해야 함.

`h` 또는 `r`이 short flag 묶음에 포함되는 경우 다음 에러 메시지를 출력하고 종료:
```
Invalid flag: -h  // 또는 -r로 대체되어야 함
Cannot use -h in a short flag bundle.

{help message}
```

모든 help message는 영어로 작성되어야 함.

- `--help` (`-h`): 명령어 사용법을 출력하고 종료. 만약 moduleName과 함께 사용되었다면 moduleName을 무시하고 help message를 출력함. help message는 다음과 같은 내용을 포함해야 함:
    - 프로젝트 설명
    - 사용법
    - positional argument 설명
    - optional argument 설명
    - flag 설명
- `--force`: 기존 go module을 덮어쓰거나 손상시킬 수 있는 작업을 강제로 실행함.
- `--docker` (`-d`): dockerfile을 생성합니다
- `--auto-migration` (`-m`): boilerplate code에서 auto migration 기능을 활성화합니다
- `--jsvm` (`-j`): jsvm 기능을 활성화합니다. pb_hooks, js migrations, js runtime을 사용 할 수 있습니다
- `--cgo-enabled`: `--docker` flag와 함께 전달되는 경우 CGO_ENABLED=1 옵션으로 build합니다.
- `--recommend` (`-r`): `--docker --auto-migration`과 동일함.

# Output Streams

- `--help`로 요청된 help message는 stdout으로 출력하고 exit code 0으로 종료함.
- 에러 발생 시 CLI가 생성하는 에러 메시지와 help message는 stderr로 출력하고 exit code 1로 종료함.
- 외부 명령 실패 시 전달하는 command output은 stderr로 출력하고 exit code 1로 종료함.
- `--force`가 제공되어 기존 go module 초기화를 계속 진행한다는 안내 메시지는 오류가 아니므로 stdout으로 출력함.

# 동작

1. 프로젝트 모듈 디렉토리에 pocketbase sdk 설치. `go get github.com/pocketbase/pocketbase@{pb-version}` 명령을 실행함. `{pb-version}`은 `--pb-version` flag로 전달된 값이 존재하면 해당 값으로, 존재하지 않으면 `latest`로 설정됨. 만약 설치에 실패했다면 명령 실행 결과로 반환된 에러 메시지를 stderr로 그대로 출력하고 종료. (exit=1)
    - `moduleName`이 전달되어 `go mod init` 실행에 실패한 경우에도 명령 실행 결과로 반환된 에러 메시지를 stderr로 그대로 출력하고 종료. (exit=1)
    - `--migration-dir` 값이 절대 경로인 경우 다음 에러 메시지를 stderr로 출력하고 종료.
        ```
        Invalid --migration-dir: absolute paths are not allowed. Use a child path relative to the PocketBase project module directory.

        {help message}
        ```
    - `--migration-dir` 값에 current directory reference(`.`)가 포함된 경우 다음 에러 메시지를 stderr로 출력하고 종료.
        ```
        Invalid --migration-dir: current directory references (`.`) are not allowed. Use a child path relative to the PocketBase project module directory.

        {help message}
        ```
    - `--migration-dir` 값에 parent directory reference(`..`)가 포함된 경우 다음 에러 메시지를 stderr로 출력하고 종료.
        ```
        Invalid --migration-dir: parent directory references (`..`) are not allowed. Use a child path relative to the PocketBase project module directory.

        {help message}
        ```
2. 지정된 flag에 따라 `templates/main.go.tmpl` 내용을 templating. 다음 template 변수가 사용됨.
    - `{{.ModulePath}}`: `go.mod` 파일에서 `module ` 뒤에 존재하는 경로로 대체되어야 함. 예를 들어 `go.mod` 파일에서 `module github.com/crmin/pb-init` 내용을 찾을 수 있을 때, `{{.ModulePath}}`는 `github.com/crmin/pb-init`로 대체되어야 함
    - `{{.MigrationDir}}`: `--migration-dir` 옵션의 값. (또는 default value `migrations`)
    - `{{.JSVMImport}}`: `--jsvm` flag가 전달되었다면 이 값을 true로 설정. 아니면 false로 설정. template에서는 `{{- if .JSVMImport}}~~import 경로~~{{- end}}` 형태로 사용 됨
    - `{{.AutoMigration}}`: `--auto-migration` flag가 전달되었다면 이 값을 `"true"`로 설정. 아니면 `"false"`로 설정. 주의: templating으로 완성된 go code에서 boolean type으로 인식되어야하므로 AutoMigration 값 자체가 boolean type이 되면 안됨
3. `--migration-dir` 디렉토리를 생성. 이 값은 slash 등으로 구분된 중첩 디렉토리 형태일 수 있음 (예를 들어 다음 값을 가질 수 있음: `migrations`, `internal/migrations`) 따라서 exist=ok 조건으로 중첩 디렉토리를 모두 생성해야 함
4. migration directory 경로에 `init.go` 파일을 생성. `templates/migration_init.go.tmpl` 파일의 내용을 templating 해서 채워넣음. 다음 template 변수가 사용됨.
    - `{{.MigrationDir}}`: `--migration-dir` 옵션의 값. (또는 default value `migrations`)
5. `--docker` flag가 전달된 경우 프로젝트 모듈 디렉토리에 `Dockerfile`, `.dockerignore` 생성. 내용은 각각 `templates/Dockerfile.tmpl`, `templates/.dockerignore.tmpl` 파일을 templating해서 사용.
    - `Dockerfile.tmpl`에서 사용되는 template variables
        - `{{.CgoEnabled}}`: `--cgo-enabled` flag가 전달된 경우 `"1"`, 아니면 `"0"`으로 설정. templating 후 정수 형태의 문자로 설정되어야 함. e.g. `CGO_ENABLED={{.CgoEnabled}}` -> `CGO_ENABLED=1`
    - `.dockerignore.tmpl`에서 사용되는 template variables
        - `{{.BinaryName}}`: module path에서 가장 마지막 경로 요소. 예를 들어 module path가 `github.com/username/app`이라면 `app`이 됨. 또는 module path가 경로 구분 없는 `test`라면 `test`가 됨. 만약 이 값이 `pocketbase`라면 이미 존재하는 값이므로 `{{.BinaryName}}`을 empty string으로 설정
6. 프로젝트 모듈 디렉토리에 `.gitignore` 파일을 생성. 내용은 `templates/.gitignore.tmpl` 템플릿을 사용. `.dockerignore.tmpl`과 같은 template variable을 사용함

# Build

templates 디렉토리의 모든 파일은 embed 되어서 build 후 binary에 포함되어야 함
