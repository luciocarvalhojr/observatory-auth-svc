# observatory-auth-svc

Go-based OIDC authentication microservice. Authenticates users via an external OIDC provider (dev: Dex), issues HS256 JWTs, and manages token revocation via Redis.

---

## Current State

- **Version**: v1.5.4 (semantic-release from `main`)
- **Pipeline**: `.github/workflows/devsecops.yml` — unified CI/CD:
  - `test-and-lint` → `security-scan` → `docker-scan` → `release` → `build-and-push` → `update-helm-chart`
  - Security gates: Trivy (CRITICAL/HIGH blocks), Gosec, Govulncheck, Gitleaks
  - Image signed with Cosign (keyless OIDC); SBOM generated (SPDX)
- **Deployed via**: External Helm chart in `luciocarvalhojr/helm-charts` (updated by pipeline on release)
- **Coverage gate**: Set to 0% (not enforced — no tests written yet)

---

## In Progress

Nothing currently in progress.

---

## Known Issues

- **No tests**: Coverage gate is explicitly set to `THRESHOLD=0` with a `TODO: raise to 80 when tests are written`. No unit or integration tests exist anywhere in the codebase.
- **No required-var validation at startup**: Config loads env vars via Viper but does not validate that required vars (OIDC_ISSUER, OIDC_CLIENT_SECRET, JWT_SECRET, REDIS_URL, etc.) are set — the service will start and fail at runtime.
- **Token revocation not persistent**: Redis blacklist is the sole revocation store. If Redis restarts, previously revoked tokens become valid again until natural expiry.
- **HS256 shared secret**: JWT signing uses a symmetric secret (`JWT_SECRET`), which means any service with the secret can forge tokens. No asymmetric alternative.

---

## Next Steps

- Write unit/integration tests and raise coverage gate to 80%
- Add startup validation for required config vars (fail fast with a clear error)
- Consider adding OpenTelemetry tracing (OTLP endpoint is configured but tracing is not implemented)

---

## Key Decisions

- **HS256 JWTs over asymmetric**: Simpler to operate; acceptable for a single-service auth boundary. Revisit if other services need to verify tokens independently.
- **Redis for state + revocation**: Stateless app instances; Redis is the shared store for OIDC CSRF state (10min TTL, single-use) and token blacklist. No database dependency.
- **Distroless runtime image**: `gcr.io/distroless/static-debian13:nonroot` — no shell, no package manager, minimal CVE surface.
- **Single unified workflow**: All CI and CD stages in one file (`devsecops.yml`); no separate release workflow.
- **Helm chart lives externally**: This repo only produces the image. The Helm chart is managed in `luciocarvalhojr/helm-charts` and updated via PR by the pipeline on each release.
- **Dex for local dev**: `docker-compose.yml` runs Dex as the OIDC provider with in-memory storage and hardcoded test users (`admin@observatory.local`, `viewer@observatory.local`, password: `password123`).
