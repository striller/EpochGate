# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in EpochGate, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please email: **security@striller.de**

Include the following in your report:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix or mitigation**: Depends on severity, typically within 2 weeks for critical issues

## Security Measures

- All dependencies are scanned for CVEs using Trivy
- REUSE compliant licensing ensures transparency
- Minimal Docker image (scratch-based) reduces attack surface
- Runs as non-root user (65534:65534)
- No external dependencies at runtime except network access

## Best Practices for Deployment

- Use HTTPS in production (reverse proxy with TLS)
- Restrict network access to EpochGate
- Keep the image updated to the latest version
- Monitor logs for blocked package attempts
