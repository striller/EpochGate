# EpochGate

A reverse proxy gatekeeper for Nexus npm repositories that enforces a minimum package age policy. Packages younger than a configurable threshold are blocked, preventing supply-chain attacks via freshly published malicious packages.

## How It Works

```
npm client → EpochGate → Nexus Registry → Upstream npm
                  ↓
          Checks package age
          via registry.npmjs.org
```

1. Client requests a package through EpochGate
2. EpochGate checks the package's last modified timestamp from the npm registry
3. If the package is younger than `MIN_AGE_DAYS`, the request is blocked with `403 Forbidden`
4. If the package passes, it is proxied to your Nexus repository
5. Results are cached in-memory for fast repeated requests

## Quick Start

### Prerequisites

- Go 1.26+
- [Nexus Repository Manager](https://www.sonatype.com/products/nexus-repository) (npm proxy repository)
- [Trivy](https://trivy.run/) (for license scanning in git hooks)
- [jq](https://jqlang.github.io/jq/) (for license check output parsing)

### Setup

```bash
# Clone the repository
git clone https://github.com/EpochGate.git
cd EpochGate

# Install git hooks
make setup-hooks

# Configure
cp .env.example .env

# Run
make run
```

### Build

```bash
make build   # Output: bin/server
```

## Configuration

Configuration is loaded from environment variables, with fallback to a `.env` file in the project root.

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_PORT` | `:8080` | Address and port to listen on |
| `NEXUS_URL` | `http://localhost:8081/repository/npm-proxy/` | Nexus npm proxy repository URL |
| `NPM_REGISTRY` | `https://registry.npmjs.org/` | Upstream npm registry for metadata |
| `MIN_AGE_DAYS` | `7` | Minimum package age in days before allowing |

Example `.env`:

```env
LISTEN_PORT=:3000
NEXUS_URL=http://nexus.internal:8081/repository/npm-proxy/
NPM_REGISTRY=https://registry.npmjs.org/
MIN_AGE_DAYS=14
```

### Configure npm client

Point your npm client to EpochGate instead of Nexus directly:

```bash
npm config set registry http://localhost:8080/
```

Or per-project in `.npmrc`:

```
registry=http://localhost:8080/
```

## Git Hooks

Git hooks are configured in `.githooks/` and activated via `make setup-hooks`. They run automatically on `git commit`.

### Pre-commit hook (`.githooks/pre-commit`)

Runs automatically when Go files or dependency files (`go.mod`, `go.sum`) are staged:

| Check | Trigger | Description |
|-------|---------|-------------|
| `go vet ./...` | Go files changed | Static analysis for common issues |
| `go build ./...` | Go files changed | Ensures code compiles |
| License scan | `go.mod`/`go.sum` changed | Blocks non-compliant licenses |

### License check (`.githooks/check-licenses`)

Uses [Trivy](https://trivy.run/) to scan all dependency licenses. Commits are blocked if any dependency uses a license outside the allowed list:

- MIT
- BSD-2-Clause
- BSD-3-Clause

Run manually:

```bash
make check-licenses
```

## Project Structure

```
EpochGate/
├── cmd/server/
│   ├── main.go           # Entrypoint, config, signal handling
│   └── server.go         # HTTP server with graceful shutdown
├── internal/
│   ├── config/
│   │   ├── config.go     # .env loading, env var resolution
│   │   └── config_test.go
│   ├── proxy/
│   │   ├── proxy.go      # Reverse proxy with age gate + cache
│   │   └── proxy_test.go
│   └── router/
│       ├── router.go     # Route definitions
│       └── router_test.go
├── .githooks/
│   ├── pre-commit        # Runs vet, build, and license checks
│   └── check-licenses    # Trivy-based license scanner
├── REUSE.toml            # REUSE copyright/license annotations
├── configs/
│   └── config.yaml       # Reference configuration
├── .env.example          # Configuration template
├── Makefile
└── go.mod
```

## Development

```bash
make run           # Start the server
make build         # Build binary to bin/server
make test          # Run all tests
make test-cover    # Run tests with coverage report
make lint          # Run golangci-lint
make reuse-lint    # Check REUSE compliance
make clean         # Remove build artifacts
```

## Docker

Build uses [semver](https://semver.org/) tags from git. Version is extracted via `git describe --tags`.

```bash
# Build (auto-tags with latest git tag, e.g. v1.2.3)
make docker-build

# Build + scan for HIGH/CRITICAL CVEs
make docker-scan

# Manual version override
VERSION=1.0.0 make docker-build

# Custom registry
REGISTRY=ghcr.io NAMESPACE=myuser make docker-build
```

**Makefile variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | latest git tag | Semver version tag |
| `REGISTRY` | `docker.io` | Docker registry host |
| `NAMESPACE` | `satriller` | Registry namespace/user |
| `IMAGE` | `epochgate` | Image name |

The pre-push hook automatically builds and scans the image when `Dockerfile` changes.

## CI/CD

Uses [Gitea Actions](https://docs.gitea.com/usage/actions/overview).

**Pipeline triggers:**
- `push` to `main` → runs tests
- `tag` (e.g. `v1.2.3`) → tests + builds linux binaries + creates Gitea release

**Docker Hub publish** (manual, requires Docker):
```bash
make docker-build    # builds image
make docker-scan     # scans for CVEs
docker push docker.io/satriller/epochgate:1.0.0
```

**Release workflow:**
```bash
git tag v1.2.3
git push origin v1.2.3
# Pipeline builds epochgate:1.2.3 + epochgate:latest
```

## Testing

```bash
make test
```

Tests use `httptest` to mock the npm registry and Nexus server. Coverage targets all packages at 100% (excluding `main()`).

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## License

This project is [REUSE](https://reuse.software/) compliant. License: [MIT](LICENSE)
