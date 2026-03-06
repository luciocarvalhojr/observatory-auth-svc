package oidc

import (
	"context"
	"fmt"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/config"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/domain"
	"golang.org/x/oauth2"
)

// Provider wraps the OIDC provider and OAuth2 config.
type Provider struct {
	provider *gooidc.Provider
	verifier *gooidc.IDTokenVerifier
	oauth2   oauth2.Config
}

// New initialises the OIDC provider by discovering the issuer metadata.
func New(ctx context.Context, cfg *config.Config) (*Provider, error) {
	provider, err := gooidc.NewProvider(ctx, cfg.OIDCIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc: discover provider %q: %w", cfg.OIDCIssuer, err)
	}

	verifier := provider.Verifier(&gooidc.Config{
		ClientID: cfg.OIDCClientID,
	})

	oauth2Config := oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.OIDCRedirectURL,
		Scopes:       []string{gooidc.ScopeOpenID, "profile", "email", "groups"},
	}

	return &Provider{
		provider: provider,
		verifier: verifier,
		oauth2:   oauth2Config,
	}, nil
}

// AuthCodeURL generates the OAuth2 authorization URL.
// state is a random value the caller must verify on callback.
func (p *Provider) AuthCodeURL(state string) string {
	return p.oauth2.AuthCodeURL(state)
}

// Exchange trades the authorization code for an ID token and
// returns the verified identity.
func (p *Provider) Exchange(ctx context.Context, code string) (*domain.Identity, error) {
	token, err := p.oauth2.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oidc: exchange code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("oidc: no id_token in response")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc: verify id_token: %w", err)
	}

	var claims struct {
		Email  string   `json:"email"`
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc: parse claims: %w", err)
	}

	return &domain.Identity{
		Subject: idToken.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
		Groups:  claims.Groups,
	}, nil
}
