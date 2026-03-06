package config

import (
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

	// Explicitly bind each env var — required for viper.Unmarshal to work
	// AutomaticEnv alone only works with viper.Get(), not Unmarshal
	viper.BindEnv("PORT")
	viper.BindEnv("OIDC_ISSUER")
	viper.BindEnv("OIDC_CLIENT_SECRET")
	viper.BindEnv("OIDC_CLIENT_ID")
	viper.BindEnv("OIDC_REDIRECT_URL")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("JWT_ACCESS_TTL")
	viper.BindEnv("REDIS_URL")
	viper.BindEnv("OTLP_ENDPOINT")
	viper.BindEnv("ENV")

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
