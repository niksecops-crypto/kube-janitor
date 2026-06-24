# Kube-Janitor: Production Deployment Guide

## Overview

Kube-Janitor runs as a long-lived Deployment in your Kubernetes cluster and continuously evicts stale finished pods and completed/failed jobs. It exposes Prometheus metrics so you can alert on errors and track cleanup throughput over time.

---

## Prerequisites

| Requirement | Minimum Version |
|-------------|----------------|
| Kubernetes  | 1.20+          |
| Helm        | 3.x            |
| Go          | 1.22+ (build only) |

---

## Quick Start (Helm)

```bash
helm upgrade --install kube-janitor ./deploy/helm/kube-janitor \
  --namespace kube-system \
  --set image.tag=latest \
  --set maxAge=48h \
  --set dryRun=false \
  --set interval=1h
```

Run in `--dry-run` mode first to preview what will be deleted:

```bash
helm upgrade --install kube-janitor ./deploy/helm/kube-janitor \
  --namespace kube-system \
  --set dryRun=true
```

---

## Configuration Reference

| Flag / Helm Value | Description | Default |
|-------------------|-------------|---------|
| `--max-age` / `maxAge` | Age threshold for finished resources | `24h` |
| `--dry-run` / `dryRun` | Preview mode — no deletions performed | `false` |
| `--interval` / `interval` | How often to run the cleanup cycle | `1h` |
| `--namespace` | Restrict to specific namespaces (repeatable) | all namespaces |
| `--metrics-addr` | Address to expose `/metrics` and `/healthz` | `:9090` |
| `--kubeconfig` | Path to kubeconfig (outside cluster only) | in-cluster config |

---

## RBAC

Kube-Janitor requires `list` and `delete` permissions on pods, jobs, and `list` on namespaces. The Helm chart creates a `ClusterRole` and `ClusterRoleBinding` automatically.

For namespace-restricted deployments (if you pass `--namespace` flags), you can downgrade to a `Role`/`RoleBinding` per namespace and remove `ClusterRole` rights.

---

## Prometheus Metrics

Kube-Janitor exposes metrics at `:9090/metrics` by default.

| Metric | Type | Description |
|--------|------|-------------|
| `kube_janitor_pods_deleted_total` | Counter | Total pods deleted since startup |
| `kube_janitor_jobs_deleted_total` | Counter | Total jobs deleted since startup |
| `kube_janitor_errors_total` | Counter | Total errors during cleanup cycles |
| `kube_janitor_cleanup_duration_seconds` | Histogram | Wall-clock time per cleanup cycle |

### Prometheus Scrape Config

```yaml
scrape_configs:
  - job_name: kube-janitor
    static_configs:
      - targets: ['kube-janitor.kube-system.svc.cluster.local:9090']
```

### Grafana Dashboard (example alerts)

```yaml
# Alert on elevated error rate
- alert: KubeJanitorErrors
  expr: increase(kube_janitor_errors_total[5m]) > 5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Kube-Janitor encountering repeated errors"

# Alert on slow cleanup cycles
- alert: KubeJanitorSlowCycle
  expr: histogram_quantile(0.95, kube_janitor_cleanup_duration_seconds_bucket) > 60
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Kube-Janitor cleanup cycles taking longer than 60s at p95"
```

---

## Production Recommendations

### 1. Start with dry-run

Before going live, run with `--dry-run=true` for one full interval and inspect the logs:

```bash
kubectl logs -n kube-system -l app.kubernetes.io/name=kube-janitor -f
```

### 2. Set a conservative max-age first

A 72h threshold is a safe starting point in most production environments — it gives engineers enough time to inspect failed pods before they're deleted.

### 3. Namespace scoping for multi-tenant clusters

In large clusters it is safer to scope Kube-Janitor to application namespaces only:

```bash
helm upgrade --install kube-janitor ./deploy/helm/kube-janitor \
  --namespace kube-system \
  --set 'namespaces={app1,app2,app3}' \
  --set maxAge=48h
```

### 4. Resource limits

The Helm chart defaults are conservative. In clusters with thousands of pods, consider increasing memory:

```yaml
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

---

## Running Outside the Cluster

```bash
go build -o kube-janitor ./cmd/janitor

./kube-janitor \
  --kubeconfig ~/.kube/config \
  --max-age 24h \
  --dry-run=true \
  --namespace default
```

---

## Troubleshooting

**Pod not being deleted even though it is old enough?**
Check that its phase is `Succeeded` or `Failed`. Kube-Janitor ignores `Running`, `Pending`, and `Unknown` pods by design.

**"failed to list pods" error in logs?**
The ServiceAccount may be missing `list` permission. Verify with:
```bash
kubectl auth can-i list pods --as=system:serviceaccount:kube-system:kube-janitor
```

**Metrics not showing up in Prometheus?**
Verify the metrics port is accessible: `kubectl port-forward svc/kube-janitor-metrics 9090:9090 -n kube-system` then `curl localhost:9090/metrics`.
