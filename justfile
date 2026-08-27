set minimum-version := '1.58.0'

[windows]
set shell := ["powershell.exe", "-NoLogo", "-Command"]
[unix]
set shell := ["sh", "-cu"]

[windows]
git_commit := env_var_or_default("GIT_COMMIT", `$c = git rev-parse --short HEAD 2>$null; if ($LASTEXITCODE -eq 0) { $c } else { "none" }`)
[unix]
git_commit := env_var_or_default("GIT_COMMIT", `git rev-parse --short HEAD 2>/dev/null || echo "none"`)

build_time := `date -u '+%Y-%m-%dT%H:%M:%SZ'`
go := env_var_or_default("GOBIN", "go")
tag := env_var_or_default("TAG", "latest")
goos := env_var_or_default("GOOS", "linux")
goarch := env_var_or_default("GOARCH", "amd64")

# show recipes list
default:
    @just --list

# build bot binary
build-bot: (_build "bot" "cmd/bot/main.go")

# build bot cli binary
build-cli: (_build "cli" "cmd/cli/main.go")

[private]
_build binary main:
    @echo "Building {{binary}} for {{goos}}/{{goarch}}..."
    GOOS={{goos}} GOARCH={{goarch}} {{go}} build -ldflags "\
        -X 'main.tag={{tag}}' \
        -X 'main.buildTime={{build_time}}' \
        -X 'main.commit={{git_commit}}'" -o {{binary}} {{main}}

# create new migration with specified name
create-migration name:
    {{go}} run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations create "{{name}}" sql
