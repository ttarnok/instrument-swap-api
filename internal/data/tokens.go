package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

const (
	ScopeRefresh        = "refresh"
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type StatefulToken struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserId    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateStatefulToken(userID int64, ttl time.Duration, scope string) (*StatefulToken, error) {

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	tokenText := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(tokenText))

	token := &StatefulToken{
		UserId:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
		Plaintext: tokenText,
		Hash:      hash[:],
	}

	return token, nil
}

func ValidateStatefulTokenPlantext(v *validator.Validator, tokenPlaintext string, label string) {
	v.Check(tokenPlaintext != "", label, "must be provided")
	v.Check(len(tokenPlaintext) == 26, label, "must be 26 bytes long")
}

type StatefulTokenModel struct {
	DB *sql.DB
}

func (m StatefulTokenModel) New(userID int64, ttl time.Duration, scope string) (*StatefulToken, error) {
	token, err := generateStatefulToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)

	return token, err
}

func (m StatefulTokenModel) Insert(token *StatefulToken) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
			VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserId, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err

}

func (m StatefulTokenModel) DeleteAllForUser(userID int64, scope string) error {
	query := `
		DELETE FROM tokens
			WHERE user_id = $1
			  AND scope = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, scope)
	return err
}
