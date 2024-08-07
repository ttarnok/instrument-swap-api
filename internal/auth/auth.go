// Package auth provides authentication functionalities for the application.
package auth

import "time"

// Claims contains the claims from a Token.
type Claims struct {
	ID      string
	Subject string
}

// TokenProvider is an interface for JWT Token functionality.
type TokenProvider interface {
	NewToken(userID int64) ([]byte, error)
	ParseClaims(token []byte) (Claims, error)
	IsValid(token []byte) bool
}

// BlacklistProvider provides functionality to blacklist tokens.
type BlacklistProvider interface {
	BlacklistToken(token string) error
	IsTokenBlacklisted(token string) (bool, error)
}

// Auth provides authentication functionality for the application.
type Auth struct {
	AccessToken    TokenProvider
	RefreshToken   TokenProvider
	BlacklistToken BlacklistProvider
}

// NewAuth a new Auth.
func NewAuth(secret string, blacklistToken BlacklistProvider) *Auth {
	return &Auth{
		AccessToken:    NewJwtTokenFactory(secret, "access", 5*time.Minute),
		RefreshToken:   NewJwtTokenFactory(secret, "refresh", 24*time.Hour),
		BlacklistToken: blacklistToken,
	}
}
