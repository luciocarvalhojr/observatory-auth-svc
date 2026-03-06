package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/config"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/domain"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/oidc"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Auth handles all authentication endpoints.
type Auth struct {
	cfg      *config.Config
	provider *oidc.Provider
	redis    *redis.Client
	ttl      time.Duration
}

// NewAuth creates a new Auth handler.
func NewAuth(cfg *config.Config, provider *oidc.Provider, rdb *redis.Client, ttl time.Duration) *Auth {
	return &Auth{cfg: cfg, provider: provider, redis: rdb, ttl: ttl}
}

// Register wires all auth routes onto the given router group.
func (a *Auth) Register(rg *gin.RouterGroup) {
	rg.GET("/login", a.Login)
	rg.GET("/callback", a.Callback)
	rg.GET("/introspect", a.Introspect)
	rg.POST("/logout", a.Logout)
}

// Login godoc
// @Summary      Initiate OIDC login
// @Description  Redirects the user to the OIDC provider login page
// @Tags         auth
// @Success      302
// @Router       /auth/login [get]
func (a *Auth) Login(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Store state in Redis for 10 minutes to validate on callback
	ctx := c.Request.Context()
	if err := a.redis.Set(ctx, stateKey(state), "1", 10*time.Minute).Err(); err != nil {
		log.Error().Err(err).Msg("failed to store state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Redirect(http.StatusFound, a.provider.AuthCodeURL(state))
}

// Callback godoc
// @Summary      OIDC callback
// @Description  Handles the OIDC provider callback, exchanges code for JWT
// @Tags         auth
// @Param        code   query  string  true  "Authorization code"
// @Param        state  query  string  true  "State"
// @Success      200    {object}  domain.TokenResponse
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Router       /auth/callback [get]
func (a *Auth) Callback(c *gin.Context) {
	ctx := c.Request.Context()

	state := c.Query("state")
	code := c.Query("code")

	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state or code"})
		return
	}

	// Verify state
	key := stateKey(state)
	if err := a.redis.Get(ctx, key).Err(); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired state"})
		return
	}
	a.redis.Del(ctx, key) // state is single-use

	identity, err := a.provider.Exchange(ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("oidc exchange failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	token, err := issueToken(identity, a.cfg.JWTSecret, a.ttl)
	if err != nil {
		log.Error().Err(err).Msg("failed to issue token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	log.Info().
		Str("subject", identity.Subject).
		Str("email", identity.Email).
		Msg("user authenticated")

	c.JSON(http.StatusOK, domain.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(a.ttl.Seconds()),
	})
}

// Introspect godoc
// @Summary      Validate a JWT
// @Description  Validates a Bearer token and returns identity claims. Used by api-gateway.
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  domain.IntrospectResponse
// @Failure      401  {object}  map[string]string
// @Router       /auth/introspect [get]
func (a *Auth) Introspect(c *gin.Context) {
	tokenStr := extractBearer(c.GetHeader("Authorization"))
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Check token blacklist
	ctx := c.Request.Context()
	if a.isRevoked(ctx, tokenStr) {
		c.JSON(http.StatusUnauthorized, domain.IntrospectResponse{Active: false})
		return
	}

	claims, err := verifyToken(tokenStr, a.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.IntrospectResponse{Active: false})
		return
	}

	c.JSON(http.StatusOK, domain.IntrospectResponse{
		Active:  true,
		Subject: claims.Subject,
		Email:   claims.Email,
	})
}

// Logout godoc
// @Summary      Revoke a JWT
// @Description  Adds the token to the blacklist until its natural expiry
// @Tags         auth
// @Security     BearerAuth
// @Success      204
// @Failure      401  {object}  map[string]string
// @Router       /auth/logout [post]
func (a *Auth) Logout(c *gin.Context) {
	tokenStr := extractBearer(c.GetHeader("Authorization"))
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := verifyToken(tokenStr, a.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Blacklist until expiry
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		ctx := c.Request.Context()
		a.redis.Set(ctx, revokedKey(tokenStr), "1", remaining)
	}

	log.Info().Str("subject", claims.Subject).Msg("user logged out")
	c.Status(http.StatusNoContent)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (a *Auth) isRevoked(ctx context.Context, token string) bool {
	err := a.redis.Get(ctx, revokedKey(token)).Err()
	return err == nil
}

func extractBearer(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func stateKey(state string) string   { return "oidc:state:" + state }
func revokedKey(token string) string { return "auth:revoked:" + token }
