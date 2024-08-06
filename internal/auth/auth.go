// Package auth provides authentication functionalities for the application.
package auth

import "time"

// Claims contains the claims from a Token.
type Claims struct {
	Subject string
}

// TokenProvider is an interface for JWT Token functionality.
type TokenProvider interface {
	NewToken(userID int64) ([]byte, error)
	ParseClaims(token []byte) (Claims, error)
}

// Auth provides authentication functionality for the application.
type Auth struct {
	AccessToken  TokenProvider
	RefreshToken TokenProvider
}

// NewAuth a new Auth.
func NewAuth(secret string) *Auth {
	return &Auth{
		AccessToken:  NewJwtTokenFactory(secret, "access", 5*time.Minute),
		RefreshToken: NewJwtTokenFactory(secret, "refresh", 24*time.Hour),
	}
}
