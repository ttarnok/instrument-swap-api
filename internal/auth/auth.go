// Package auth provides authentication functionalities for the application.
package auth

// Claims contains the claims from an Access Token.
type Claims struct {
	Subject string
}

// AccessTokenProvider is an interface for AccessToken functionality.
type AccessTokenProvider interface {
	New(userID int64) ([]byte, error)
	ParseClaims(token []byte) (Claims, error)
}

// Auth provides authentication functionality for the application.
type Auth struct {
	AccessToken AccessTokenProvider
}

// NewAuth a new Auth.
func NewAuth(secret string) *Auth {
	return &Auth{AccessToken: NewAccessToken(secret)}
}
