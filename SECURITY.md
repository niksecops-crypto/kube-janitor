# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.x     | ✅        |
| < 1.0   | ❌        |

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report vulnerabilities privately via GitHub Security Advisories:
👉 [Report a vulnerability](https://github.com/niksecops-crypto/kube-janitor/security/advisories/new)

Or email: **security@niksecops.dev**

Include:
- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact

You will receive an acknowledgement within **48 hours** and a resolution timeline within **7 days**.

## Security Considerations

- kube-janitor requires ClusterRole with `delete` permissions on pods and jobs
- Always run with the principle of least privilege — scope to specific namespaces via `--namespace` when possible
- The container runs as `nobody` (UID 65534) with a read-only filesystem
