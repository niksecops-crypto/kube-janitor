# Kube-Janitor: Automated Kubernetes Resource Cleanup

[![CI](https://github.com/niksecops-crypto/kube-janitor/actions/workflows/ci.yml/badge.svg)](https://github.com/niksecops-crypto/kube-janitor/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/niksecops-crypto/kube-janitor)](https://goreportcard.com/report/github.com/niksecops-crypto/kube-janitor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/GHCR-ghcr.io-blue?logo=docker)](https://github.com/niksecops-crypto/kube-janitor/pkgs/container/kube-janitor)

Kube-Janitor is a simple, lightweight utility for cleaning up old or unused resources in your Kubernetes cluster. It's designed for DevOps engineers who want to keep their clusters clean and reduce resource fragmentation.

## Features

- **Automated Pod Cleanup**: Deletes finished (Succeeded/Failed) pods that are older than a specified threshold.
- **Automated Job Cleanup**: Cleans up completed or failed Batch Jobs and their associated pods.
- **Dry-run Mode**: Preview deletions before they happen to avoid accidental data loss.
- **Configurable Thresholds**: Set how long to keep finished resources.
- **In-Cluster & Out-of-Cluster Support**: Works inside your cluster or as a standalone CLI tool.

## Why Kube-Janitor?

Kubernetes doesn't always clean up finished pods and jobs immediately, leading to thousands of "stale" resources that clutter `kubectl get pods` and increase pressure on the API server. Kube-Janitor automates this housekeeping.

## Getting Started

### Prerequisites

- Kubernetes cluster (1.20+)
- Go 1.22+ (for building from source)
- `kubectl` configured to your cluster

### Quick Start (Local)

```bash
# Clone the repository
git clone https://github.com/niksecops-crypto/kube-janitor.git
cd kube-janitor

# Build and run with 24-hour cleanup threshold and dry-run mode
go run cmd/janitor/main.go --max-age=24h --dry-run=true
```

### Deploy to Kubernetes (Helm)

We provide a Helm chart for easy deployment as a background service.

```bash
helm upgrade --install kube-janitor deploy/helm/kube-janitor \
  --namespace kube-system \
  --set maxAge=48h \
  --set dryRun=false
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--max-age` | Maximum age of finished resources before deletion | `24h` |
| `--dry-run` | If true, do not perform actual deletions | `false` |
| `--interval`| How often to run the cleanup cycle | `1h` |
| `--kubeconfig` | Path to kubeconfig (if running outside cluster) | In-cluster config |

## Production Best Practices

- Always run in `dry-run` mode first to verify which resources will be deleted.
- Set an appropriate `max-age` (e.g., `48h` or `72h`) to allow for manual inspection of failed pods.
- Run in a dedicated namespace with appropriate RBAC permissions.

## Documentation

- [Production Deployment Guide](docs/production-guide.md) — Helm deployment, Prometheus alerting, RBAC, troubleshooting

## License

Distributed under the MIT License. See `LICENSE` for more information.

---
*Maintained by [niksecops-crypto](https://github.com/niksecops-crypto)*
