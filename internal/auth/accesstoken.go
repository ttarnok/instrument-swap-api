package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/pascaldekloe/jwt"
)

// AccessToken provides functionality for handling access tokens.
type AccessToken struct {
	secret string
}

// NewAccessToken creates a new AccessToken value.
func NewAccessToken(secret string) *AccessToken {
	return &AccessToken{secret: secret}
}

// New creates a new access token or returns an error.
func (a *AccessToken) New(userID int64) ([]byte, error) {
	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(userID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "instrument-swap.example.example"
	claims.Audiences = []string{"instrument-swap.example.example"}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(a.secret))

	if err != nil {
		return nil, err
	}
	return jwtBytes, nil
}

// ParseClaims validates an Acces token.
// Returns an error if the token is not valid.
func (a *AccessToken) ParseClaims(token []byte) (Claims, error) {

	claims, err := jwt.HMACCheck([]byte(token), []byte(a.secret))

	if err != nil {
		return Claims{}, err
	}

	if !claims.Valid(time.Now()) {
		return Claims{}, errors.New("expired token")
	}

	if claims.Issuer != "instrument-swap.example.example" {
		return Claims{}, errors.New("invalid issuer")
	}

	if !claims.AcceptAudience("instrument-swap.example.example") {
		return Claims{}, errors.New("invalid aqccepted audience")
	}

	return Claims{Subject: claims.Subject}, nil
}
