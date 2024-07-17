package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// Swap represents an instrument swap record in the application.
type Swap struct {
	ID                    int64      `json:"id"`
	CreatedAt             time.Time  `json:"created_at"`
	RequesterInstrumentID int64      `json:"requester_instrument_id"`
	RecipientInstrumentID int64      `json:"recipient_instrument_id"`
	IsAccepted            bool       `json:"is_accepted"`
	AcceptedAt            *time.Time `json:"accepted_at"`
	IsRejected            bool       `json:"is_rejected"`
	RejectedAt            *time.Time `json:"rejected_at"`
	IsEnded               bool       `json:"is_ended"`
	EndedAt               *time.Time `json:"ended_at"`
	Version               int32      `json:"version"`
}

// ValidateSwap checks the validity of a swap,
// adds all found validation errors into the validator.
func ValidateSwap(v *validator.Validator, swap *Swap) {
	v.Check(swap.RequesterInstrumentID != 0, "requester_instrument_id", "must not be empty")
	v.Check(swap.RequesterInstrumentID >= 0, "requester_instrument_id", "must be greater than 0")

	v.Check(swap.RecipientInstrumentID != 0, "recipient_instrument_id", "must not be empty")
	v.Check(swap.RecipientInstrumentID >= 0, "recipient_instrument_id", "must be greater than 0")
}

// SwapModel represents the database layer and provides functionality to interact with the database.
type SwapModel struct {
	DB *sql.DB
}

// GetAllForUser retrieves all swaps related to the given user.
// Returns an error if the retrieval is not possible.
func (s *SwapModel) GetAllForUser(userID int64) (swaps []*Swap, err error) {

	query := `
	SELECT id, created_at, requester_instrument_id, recipient_instrument_id, is_accepted,
		accepted_at, is_rejected, rejected_at, is_ended, ended_at, version
	FROM (
			SELECT s.*, irec.owner_user_id irec_owner_user_id, ireq.owner_user_id ireq_owner_user_id
			FROM swaps s
			JOIN instruments ireq ON s.requester_instrument_id = ireq.id
			JOIN instruments irec ON s.recipient_instrument_id = irec.id
			)
	WHERE irec_owner_user_id = $1
		 OR ireq_owner_user_id = $2
	ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}

	defer func() {
		errClose := rows.Close()
		if err == nil {
			err = errClose
		}
	}()

	swaps = []*Swap{}

	for rows.Next() {

		var swap Swap

		err := rows.Scan(
			&swap.ID,
			&swap.CreatedAt,
			&swap.RequesterInstrumentID,
			&swap.RecipientInstrumentID,
			&swap.IsAccepted,
			&swap.AcceptedAt,
			&swap.IsRejected,
			&swap.RejectedAt,
			&swap.IsEnded,
			&swap.EndedAt,
			&swap.Version,
		)

		if err != nil {
			return nil, err
		}

		swaps = append(swaps, &swap)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return swaps, nil
}

// Get retrieves a swap based on the given swap id.
// Returns ErrRecordNotFound an error if the retrieval is not possible.
func (s *SwapModel) Get(id int64) (*Swap, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, requester_instrument_id, recipient_instrument_id, is_accepted,
			accepted_at, is_rejected, rejected_at, is_ended, ended_at, version
		FROM swaps
		WHERE id = $1
		ORDER BY id`

	var swap Swap

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&swap.ID,
		&swap.CreatedAt,
		&swap.RequesterInstrumentID,
		&swap.RecipientInstrumentID,
		&swap.IsAccepted,
		&swap.AcceptedAt,
		&swap.IsRejected,
		&swap.RejectedAt,
		&swap.IsEnded,
		&swap.EndedAt,
		&swap.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &swap, nil
}

// GetByInstrumentID returns swaps based on an instrument id.
// Returns an error if the retrieval is not possible.
func (s *SwapModel) GetByInstrumentID(id int64) (*Swap, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, requester_instrument_id, recipient_instrument_id, is_accepted,
			accepted_at, is_rejected, rejected_at, is_ended, ended_at, version
		FROM swaps
		WHERE (requester_instrument_id = $1 OR recipient_instrument_id = $2)
		ORDER BY id`

	var swap Swap

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, id, id).Scan(
		&swap.ID,
		&swap.CreatedAt,
		&swap.RequesterInstrumentID,
		&swap.RecipientInstrumentID,
		&swap.IsAccepted,
		&swap.AcceptedAt,
		&swap.IsRejected,
		&swap.RejectedAt,
		&swap.IsEnded,
		&swap.EndedAt,
		&swap.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &swap, nil
}

func (s *SwapModel) Insert(swap *Swap) error {

	query := `
		INSERT INTO swaps (requester_instrument_id, recipient_instrument_id)
			VALUES($1, $2)
		RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.DB.
		QueryRowContext(ctx, query, swap.RequesterInstrumentID, swap.RecipientInstrumentID).
		Scan(&swap.ID, &swap.CreatedAt, &swap.Version)
}

func (s *SwapModel) Update(swap *Swap) error {

	query := `
		UPDATE swaps
			SET is_accepted = $1,
					accepted_at = $2,
					is_rejected = $3,
					rejected_at = $4,
					is_ended = $5,
					ended_at = $6,
					version = version + 1
		WHERE id = $7
		  AND version = $8
		RETURNING version`

	args := []any{
		swap.IsAccepted,
		swap.AcceptedAt,
		swap.IsRejected,
		swap.RejectedAt,
		swap.IsEnded,
		swap.EndedAt,
		swap.ID,
		swap.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&swap.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
