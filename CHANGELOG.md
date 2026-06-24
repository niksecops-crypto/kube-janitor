# Changelog

All notable changes to kube-janitor are documented here.

## [1.1.0] - 2024-12-10
### Added
- Multi-arch Docker images (linux/amd64, linux/arm64) via GHCR
- GitHub Actions CI: test, lint, release workflow
- Makefile with `build`, `test`, `lint`, `docker`, `push` targets
- Dockerfile using distroless base (minimal attack surface)
- Unit tests with `k8s.io/client-go/kubernetes/fake` (~80% coverage)

### Changed
- JSON structured logging via `log/slog` (was plain `log`)

## [1.0.0] - 2024-10-15
### Added
- Initial release: automated cleanup of Succeeded/Failed pods
- Batch job cleanup with cascading pod deletion
- Dry-run mode (`--dry-run`)
- In-cluster and out-of-cluster kubeconfig support
- RBAC manifests and Helm chart
