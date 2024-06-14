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
	RequesterInstrumentId int64      `json:"requester_instrument_id"`
	RecipientInstrumentId int64      `json:"recipient_instrument_id"`
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
	v.Check(swap.RequesterInstrumentId != 0, "requester_instrument_id", "must not be empty")
	v.Check(swap.RequesterInstrumentId >= 0, "requester_instrument_id", "must be greater than 0")

	v.Check(swap.RecipientInstrumentId != 0, "recipient_instrument_id", "must not be empty")
	v.Check(swap.RecipientInstrumentId >= 0, "recipient_instrument_id", "must be greater than 0")
}

type SwapModel struct {
	DB *sql.DB
}

func (s SwapModel) GetAll() ([]*Swap, error) {

	query := `
		SELECT id, created_at, requester_instrument_id, recipient_instrument_id, is_accepted,
			accepted_at, is_rejected, rejected_at, is_ended, ended_at, version
		FROM swaps
		ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	swaps := []*Swap{}

	for rows.Next() {

		var swap Swap

		err := rows.Scan(
			&swap.ID,
			&swap.CreatedAt,
			&swap.RequesterInstrumentId,
			&swap.RecipientInstrumentId,
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

func (s SwapModel) Get(id int64) (*Swap, error) {

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
		&swap.RequesterInstrumentId,
		&swap.RecipientInstrumentId,
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

func (s SwapModel) GetByInstrumentId(id int64) (*Swap, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, requester_instrument_id, recipient_instrument_id, is_accepted,
			accepted_at, is_rejected, rejected_at, is_ended, ended_at, version
		FROM swaps
		WHERE (requester_instrument_id = $1 OR recipient_instrument_id = $2)
		  AND is_ended = FALSE`

	var swap Swap

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, id, id).Scan(
		&swap.ID,
		&swap.CreatedAt,
		&swap.RequesterInstrumentId,
		&swap.RecipientInstrumentId,
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

func (s SwapModel) Insert(swap *Swap) error {

	query := `
		INSERT INTO swaps (requester_instrument_id, recipient_instrument_id)
			VALUES($1, $2)
		RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.DB.
		QueryRowContext(ctx, query, swap.RequesterInstrumentId, swap.RecipientInstrumentId).
		Scan(&swap.ID, &swap.CreatedAt, &swap.Version)
}
