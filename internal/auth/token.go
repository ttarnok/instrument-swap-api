package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pascaldekloe/jwt"
)

// JwtTokenFactory provides functionality for creating and parsing jwt tokens.
type JwtTokenFactory struct {
	secret     string
	tokenType  string
	expiration time.Duration
}

// NewJwtTokenFactory creates a new Token with the specified type.
func NewJwtTokenFactory(secret string, tokenType string, expiration time.Duration) *JwtTokenFactory {
	return &JwtTokenFactory{secret: secret, tokenType: tokenType, expiration: expiration}
}

// NewToken creates a new jwt token or returns an error.
func (t *JwtTokenFactory) NewToken(userID int64) ([]byte, error) {
	var claims jwt.Claims
	claims.ID = uuid.NewString()
	claims.Subject = strconv.FormatInt(userID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(t.expiration))
	claims.Issuer = "instrument-swap.example.example"
	claims.Audiences = []string{"instrument-swap.example.example"}
	claims.Set = map[string]interface{}{"token_type": t.tokenType}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(t.secret))

	if err != nil {
		return nil, err
	}
	return jwtBytes, nil
}

// IsValid returns whether the claims set may be accepted for processing at the moment of execution.
func (t *JwtTokenFactory) IsValid(token []byte) bool {

	claims, err := jwt.HMACCheck([]byte(token), []byte(t.secret))
	if err != nil {
		return false
	}

	return claims.Valid(time.Now())
}

// ParseClaims validates a token.
// Returns an error if the token is not valid.
func (t *JwtTokenFactory) ParseClaims(token []byte) (Claims, error) {

	claims, err := jwt.HMACCheck([]byte(token), []byte(t.secret))

	if err != nil {
		return Claims{}, err
	}

	tokenType, ok := claims.Set["token_type"].(string)
	if !ok {
		return Claims{}, errors.New("should contain token_type claim")
	}

	if tokenType != t.tokenType {
		return Claims{}, errors.New("incorrect token type")
	}

	if claims.Issuer != "instrument-swap.example.example" {
		return Claims{}, errors.New("invalid issuer")
	}

	if !claims.AcceptAudience("instrument-swap.example.example") {
		return Claims{}, errors.New("invalid aqccepted audience")
	}

	return Claims{ID: claims.ID, Subject: claims.Subject}, nil
}
