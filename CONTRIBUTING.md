# Contributing to kube-janitor

## Getting Started

```bash
git clone https://github.com/niksecops-crypto/kube-janitor.git
cd kube-janitor
go mod download
make test
```

## Development Workflow

1. Fork the repo and create a branch: `git checkout -b fix/my-fix`
2. Make changes, add tests
3. Run `make test` — all tests must pass
4. Run `make lint` — no lint errors
5. Open a PR against `main`

## Running Tests

```bash
make test          # all tests with race detector
make test-short    # fast subset, no network
```

Tests use `k8s.io/client-go/kubernetes/fake` — no real cluster needed.

## Code Style

- Follow standard Go conventions (`gofmt`, `golangci-lint`)
- Structured JSON logging via `log/slog` — no `fmt.Print` in production paths
- New features need tests; bug fixes need a regression test

## Reporting Issues

Open a [GitHub Issue](https://github.com/niksecops-crypto/kube-janitor/issues) with:
- Kubernetes version
- kube-janitor version (`--version`)
- Minimal reproduction steps
