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

```bash
# Copy and fill in your values
cp .env.example .env

go run cmd/api/main.go
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
