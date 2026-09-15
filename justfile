# https://just.systems
default:
    just --choose

run:
    go run main.go

build:
    go build -ldflags="-s -w" -trimpath -o build/cliptrans .

update-and-check: update check

check: tidy-diff check-parallel

[parallel]
check-parallel: oxfmt lint test-coverage actionlint goreleaser-check

update: get-go get-toolchain get-deps tidy gomajor

get-go:
    go get go@latest

get-toolchain:
    go get toolchain@latest

get-deps:
    go get -u -t ./...

tidy:
    go mod tidy -v

tidy-diff:
    go mod tidy -diff

gomajor: # https://github.com/icholy/gomajor
    gomajor list

lint:
    golangci-lint run

lint-fix:
    golangci-lint run --fix

fmt:
    golangci-lint fmt

test:
    go test -race ./... -coverprofile=./cover.out -covermode=atomic -coverpkg=./...

test-coverage: test
    go-test-coverage --config=./.testcoverage.yml

test-integration:
    go test -tags=integration ./...

govulncheck: # go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...

coverage-html:
    go tool cover -html=cover.out

coverage-func:
    go tool cover -func=cover.out

scc: # https://github.com/boyter/scc
    scc

tokei: #https://github.com/xampprocky/tokei
    tokei

next:
    svu next --always

tag:
    git tag $(svu next --always)

prek-check-all:
    prek run --all-files

prek-update:
    prek update

go-install:
    go install github.com/jeduden/mdsmith/cmd/mdsmith@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest

mdsmith-install: # go install github.com/jeduden/mdsmith/cmd/mdsmith@latest
    go install github.com/jeduden/mdsmith/cmd/mdsmith@latest

mdsmith: # https://github.com/jeduden/mdsmith
    mdsmith check

go-size-analyzer-install: # https://github.com/Zxilly/go-size-analyzer
    go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest

go-size-analyzer: build
    gsa --tui --verbose build/cliptrans

go-size-analyzer-web: build
    gsa --web --verbose build/cliptrans

codegraph-init:
    codegraph init

actionlint-install: # https://github.com/rhysd/actionlint
    go install github.com/rhysd/actionlint/cmd/actionlint@latest

actionlint: # https://github.com/rhysd/actionlint
    actionlint

go-test-coverage-install: # https://github.com/vladopajic/go-test-coverage
    go install github.com/vladopajic/go-test-coverage/v2@latest

oxfmt:
    oxfmt

goreleaser-check:
    goreleaser check

goreleaser-snapshot: # https://github.com/goreleaser/goreleaser
    goreleaser release --snapshot --clean

depth-install: # https://github.com/KyleBanks/depth
    go install github.com/KyleBanks/depth/cmd/depth@latest

depth: # https://github.com/KyleBanks/depth
    depth .

# Validate and Run GitHub Actions locally.
wrkflw: # https://github.com/bahdotsh/wrkflw
    wrkflw

# https://docs.zizmor.sh/quickstart/
zizmor:
    zizmor

roborev-ui: # https://github.com/kenn-io/roborev
    roborev ui

roborev-tui: # https://github.com/kenn-io/roborev
    roborev tui

review:
    ocr
