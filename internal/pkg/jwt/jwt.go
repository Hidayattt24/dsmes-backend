// Package jwt provides JWT token generation, signing, and claims extraction.
//
// This package is intentionally separate from the JWT middleware
// (internal/middleware/jwt.go). The middleware verifies incoming tokens on
// protected routes; this package is used by the auth service to ISSUE tokens
// on successful login.
//
// Token strategy:
//   - Access Token:  short-lived (15 min default), sent in Authorization header.
//   - Refresh Token: long-lived (7 days default), used to obtain new access tokens.
//
// Both tokens use HS256 (HMAC-SHA256) with the shared JWT_SECRET from config.
// For higher security in multi-service environments, consider RS256 (RSA keys).
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/dsmes/dsmes-backend/internal/config"
)

// Claims defines the payload embedded in every JWT issued by this application.
// Using a custom struct (not jwt.MapClaims) gives type-safe claim access.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // admin | staff | user
	jwt.RegisteredClaims
}

// TokenPair holds both tokens returned after a successful authentication.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"` // Unix timestamp of access token expiry
}

// Manager handles token generation and parsing.
// One instance is created during bootstrap and injected into auth services.
type Manager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// NewManager creates a JWT Manager from application config.
// NewManager creates a JWT Manager from application config.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		secret:          []byte(cfg.JWT.Secret),
		accessTokenTTL:  cfg.JWT.AccessTokenTTL,
		refreshTokenTTL: cfg.JWT.RefreshTokenTTL,
		issuer:          cfg.JWT.Issuer,
	}
}

// GenerateTokenPair creates a new access + refresh token pair for the given user.
// userID should be the UUID primary key; role is "admin", "staff", or "user".
func (m *Manager) GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(m.accessTokenTTL)

	// ── Access token ──────────────────────────────────────────────────────────
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			ID:        uuid.NewString(), // jti — unique token identifier
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to sign access token: %w", err)
	}

	// ── Refresh token ─────────────────────────────────────────────────────────
	refreshClaims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenTTL)),
			ID:        uuid.NewString(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString(m.secret)
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}

// ParseAccessToken validates and parses an access token string.
// Returns the embedded Claims or an error if the token is invalid/expired.
func (m *Manager) ParseAccessToken(tokenString string) (*Claims, error) {
	return m.parse(tokenString)
}

// ParseRefreshToken validates and parses a refresh token string.
func (m *Manager) ParseRefreshToken(tokenString string) (*Claims, error) {
	return m.parse(tokenString)
}

// parse is the internal token validation and parsing implementation.
func (m *Manager) parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		// Validate that the signing method is HS256 — reject tokens signed with other algorithms.
		// This prevents the "alg: none" attack.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))

	if err != nil {
		return nil, fmt.Errorf("jwt: invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("jwt: token is not valid")
	}

	return claims, nil
}
