package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// Generic user errors.
// These errors can be tested using errors.Is.
var (
	ErrDuplicateEmail = errors.New("duplicate email") // "duplicate email"
)

// AnonymousUser represents an unauthenticated user.
var AnonymousUser = &User{}

// User struct represents an user.
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// IsAnonymous returns true if the given user is AnonymousUser.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// password is and internal type to help dealing with passwords.
type password struct {
	plaintext *string
	hash      []byte
}

// Set sets and encrypts the given password string.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches compares the given password string with the stored encrypted password.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// ValidateEmail validates the given email address.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email addrerss")
}

// ValidatePasswordPlaintext validates the format of the given password.
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

// ValidateUser validates the fields of the given user.
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// UserModel represents the user model, that stores users in a database.
type UserModel struct {
	DB *sql.DB
}

// Insert inserts the given user into the database.
// Returns ErrDuplicateEmail if the given email is already stored.
func (m *UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
			VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetAll retrieves all users from the database.
func (m *UserModel) GetAll() (users []*User, err error) {
	query := `
	SELECT id, created_at, name, email, password_hash, activated, version
		FROM users
	WHERE is_deleted = FALSE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		errClose := rows.Close()
		if err == nil {
			err = errClose
		}
	}()

	users = []*User{}

	for rows.Next() {

		var user User

		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.Name,
			&user.Email,
			&user.Password.hash,
			&user.Activated,
			&user.Version,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil

}

// GetByEmail retrieves the user from the database with the given email.
// Returns ErrRecordNotFound if the user is not found.
func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
			FROM users
		WHERE email = $1
			AND is_deleted = FALSE`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// GetByID retrieves the user from the database with the given user id.
// Returns ErrRecordNotFound if the user is not found.
func (m *UserModel) GetByID(id int64) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
			FROM users
		WHERE id = $1
		  AND is_deleted = FALSE`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update updates the given user in the databse.
// Returns ErrDuplicateEmail if the given email is already stored in the database.
// Returns ErrEditConflict in case of conflicting update.
func (m *UserModel) Update(user *User) error {
	query := `
		UPDATE users
			SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		WHERE id = $5
			AND version = $6
			AND is_deleted = FALSE
		RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Delete deletes the user from the database with the given user id.
// Returns ErrRecordNotFound if the given user is not found.
func (m *UserModel) Delete(id int64) error {

	if id < 0 {
		return ErrRecordNotFound
	}

	query := `
		UPDATE users
			SET is_deleted = true, deleted_at = NOW()
		WHERE id = $1
		  AND is_deleted = FALSE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}
