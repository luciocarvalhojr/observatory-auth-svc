package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for auth-svc.
// Values are loaded from environment variables.
type Config struct {
	Port string `mapstructure:"PORT"`

	// OIDC
	OIDCIssuer      string `mapstructure:"OIDC_ISSUER"`
	OIDCClientID    string `mapstructure:"OIDC_CLIENT_ID"`
	OIDCRedirectURL string `mapstructure:"OIDC_REDIRECT_URL"`

	// JWT
	JWTSecret    string `mapstructure:"JWT_SECRET"`
	JWTAccessTTL string `mapstructure:"JWT_ACCESS_TTL"`

	// Redis
	RedisURL string `mapstructure:"REDIS_URL"`

	// OpenTelemetry
	OTLPEndpoint string `mapstructure:"OTLP_ENDPOINT"`

	// App
	Env string `mapstructure:"ENV"`
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8081")
	viper.SetDefault("JWT_ACCESS_TTL", "15m")
	viper.SetDefault("ENV", "production")
	viper.SetDefault("OTLP_ENDPOINT", "http://jaeger:4318")

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
