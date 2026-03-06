# observatory-auth-svc

OIDC-based authentication service for the [Observatory](https://github.com/luciocarvalhojr/observatory) platform.

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/auth/login` | Redirect to OIDC provider |
| `GET` | `/auth/callback` | OIDC callback → issue JWT |
| `GET` | `/auth/introspect` | Validate Bearer token |
| `POST` | `/auth/logout` | Revoke token |
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe |

## Local Development

The easiest way to get all dependencies (Redis and a local OIDC provider) up and running is using Docker Compose.

```bash
docker compose up --build
```

Alternatively, you can run the service manually:

1.  **Start dependencies:** `docker compose up dex redis -d`
2.  **Copy and fill in your values:** `cp .env.example .env` (if applicable)
3.  **Run:** `go run cmd/api/main.go`

## Local Validation (Docker Compose)

The `docker-compose.yml` provides a complete environment including **Dex** (simulated OIDC provider) and **Redis**.

### 1. Start the Environment
```bash
docker compose up --build
```

### 2. Configure Local Hostname (Optional but Recommended)
The local OIDC issuer is configured as `http://dex:5556/dex`. For your browser to follow redirects correctly during the login flow, add `dex` to your local hosts file:

```bash
echo "127.0.0.1 dex" | sudo tee -a /etc/hosts
```

### 3. Test the Auth Flow
1.  Navigate to `http://localhost:8081/auth/login`.
2.  You will be redirected to the Dex login page.
3.  Log in using one of the pre-configured test users:

| Email | Password | Role/Group |
|---|---|---|
| `admin@observatory.local` | `password123` | `admin` |
| `viewer@observatory.local` | `password123` | `viewer` |

4.  After successful login, you will be redirected back to the service, which will issue a JWT.

### 4. Verify Health & Metrics
```bash
curl http://localhost:8081/healthz
curl http://localhost:8081/readyz
```

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `PORT` | Listen port | `8081` |
| `OIDC_ISSUER` | OIDC provider URL | required |
| `OIDC_CLIENT_ID` | OIDC client ID | required |
| `OIDC_REDIRECT_URL` | Callback URL | required |
| `JWT_SECRET` | Signing key | required |
| `JWT_ACCESS_TTL` | Token TTL | `15m` |
| `REDIS_URL` | Redis URL | required |
| `OTLP_ENDPOINT` | Jaeger OTLP endpoint | `http://jaeger:4318` |
| `ENV` | `development` or `production` | `production` |

## Run Tests

```bash
go test -v -race -cover ./...
```

## Build Validation (Docker)

Ensure the container build process is correct:

```bash
docker build -t auth-svc .
```

## Deploy

```bash
helm install auth-svc ./helm \
  --namespace observatory \
  --create-namespace \
  --set env.OIDC_ISSUER=https://your-idp \
  --set env.OIDC_CLIENT_ID=your-client-id \
  --set secrets.JWT_SECRET=your-secret \
  --set secrets.REDIS_URL=redis://redis:6379
```
