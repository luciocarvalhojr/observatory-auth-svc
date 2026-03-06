package domain

import "github.com/golang-jwt/jwt/v5"

// Identity represents a verified OIDC user.
type Identity struct {
	Subject string
	Email   string
	Name    string
	Groups  []string
}

// Claims represents the JWT payload issued by auth-svc.
type Claims struct {
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
	jwt.RegisteredClaims
}

// TokenResponse is returned to the caller after successful auth.
type TokenResponse struct {
	AccessToken string `json:"access_token"` //#nosec G117 -- intentional: this is the API response body, not a hardcoded secret
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

// IntrospectResponse is returned by the token validation endpoint.
type IntrospectResponse struct {
	Active  bool   `json:"active"`
	Subject string `json:"sub,omitempty"`
	Email   string `json:"email,omitempty"`
}
