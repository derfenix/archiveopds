{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:

{
  # https://devenv.sh/basics/
  env.GREET = "archiveopds";

  # https://devenv.sh/packages/
  packages = with pkgs; [
    git
    golangci-lint
    gopls
    gotools
    gofumpt
    delve
    govulncheck
    goreleaser
  ];

  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/scripts/
  scripts.build.exec = "make build";
  scripts.test.exec = "go test ./...";
  scripts.testrace.exec = "go test -race -count=1 ./...";
  scripts.cover.exec = "go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -n 1";
  scripts.lint.exec = "golangci-lint run";
  scripts.fmt.exec = "golangci-lint fmt";
  scripts.vet.exec = "go vet ./...";
  scripts.tidy.exec = "go mod tidy && go mod verify";
  scripts.run.exec = "go run ./cmd/archiveopds serve";
  scripts.clean.exec = "make clean";

  # https://devenv.sh/basics/
  enterShell = ''
    echo "📚 archiveopds devenv"
    echo "   Go: $(go version | cut -d' ' -f3)"
    echo ""
    echo "   tasks:"
    echo "     build      — make build (bin/archiveopds)"
    echo "     run        — go run ./cmd/archiveopds serve"
    echo "     test       — go test ./..."
    echo "     testrace   — go test -race -count=1 ./..."
    echo "     cover      — тесты + суммарное покрытие"
    echo "     lint       — golangci-lint run"
    echo "     fmt        — golangci-lint fmt"
    echo "     vet        — go vet ./..."
    echo "     tidy       — go mod tidy && go mod verify"
    echo "     clean      — remove binary"
  '';

  # https://devenv.sh/tests/
  enterTest = ''
    go version
    go build ./...
    go test ./...
  '';

  # https://devenv.sh/git-hooks/
  git-hooks.hooks = {
    gofmt.enable = true;
    govet.enable = true;
    golangci-lint.enable = true;
  };

  # See full reference at https://devenv.sh/reference/options/
}
