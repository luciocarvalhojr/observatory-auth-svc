module github.com/luciocarvalhojr/observatory-auth-svc

go 1.24

require (
	github.com/coreos/go-oidc/v3 v3.11.0
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/redis/go-redis/v9 v9.7.0
	github.com/rs/zerolog v1.33.0
	github.com/spf13/viper v1.19.0
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.4
	go.opentelemetry.io/otel v1.33.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.33.0
	go.opentelemetry.io/otel/sdk v1.33.0
	golang.org/x/oauth2 v0.25.0
)
