package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for auth-svc.
// Values are loaded from environment variables.
type Config struct {
	Port string `mapstructure:"PORT"`

	// OIDC
	OIDCIssuer       string `mapstructure:"OIDC_ISSUER"`
	OIDCClientID     string `mapstructure:"OIDC_CLIENT_ID"`
	OIDCRedirectURL  string `mapstructure:"OIDC_REDIRECT_URL"`
	OIDCClientSecret string `mapstructure:"OIDC_CLIENT_SECRET"`

	// JWT
	JWTSecret    string `mapstructure:"JWT_SECRET" json:"-"`
	JWTAccessTTL string `mapstructure:"JWT_ACCESS_TTL"`

	// Redis
	RedisURL string `mapstructure:"REDIS_URL"`

	// OpenTelemetry
	OTLPEndpoint string `mapstructure:"OTLP_ENDPOINT"`

	// App
	Env string `mapstructure:"ENV"`
}

func mustBindEnv(keys ...string) {
	for _, key := range keys {
		if err := viper.BindEnv(key); err != nil {
			panic(fmt.Sprintf("viper.BindEnv(%q): %v", key, err))
		}
	}
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	// Explicitly bind each env var — required for viper.Unmarshal to work
	// AutomaticEnv alone only works with viper.Get(), not Unmarshal
	mustBindEnv(
		"PORT",
		"OIDC_ISSUER",
		"OIDC_CLIENT_ID",
		"OIDC_CLIENT_SECRET",
		"OIDC_REDIRECT_URL",
		"JWT_SECRET",
		"JWT_ACCESS_TTL",
		"REDIS_URL",
		"OTLP_ENDPOINT",
		"ENV",
	)

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
