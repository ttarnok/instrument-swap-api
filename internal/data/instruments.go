package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// Instrument represents an instrument record in the apprication.
type Instrument struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"-"`
	Name            string    `json:"name"`
	Manufacturer    string    `json:"manufacturer"`
	ManufactureYear int32     `json:"manufacture_year"`
	Type            string    `json:"type"`
	EstimatedValue  int64     `json:"estimated_value"`
	Condition       string    `json:"condition"`
	Description     string    `json:"description"`
	FamousOwners    []string  `json:"famous_owners"`
	OwnerUserId     int64     `json:"owner_user_id"`
	IsSwapped       bool      `json:"is_swapped"`
	Version         int32     `json:"version"`
}

// ValidateInstrument checks the validity of an Instrument,
// adds all found validtaion errors into the validator.
func ValidateInstrument(v *validator.Validator, instrument *Instrument) {
	v.Check(instrument.Name != "", "name", "must be provided")
	v.Check(len(instrument.Name) <= 500, "name", "must not be more than 500 bytes long")

	v.Check(instrument.Name != "", "manufacturer", "must be provided")
	v.Check(len(instrument.Name) <= 500, "manufacturer", "must not be more than 500 bytes long")

	v.Check(instrument.ManufactureYear != 0, "manufacture_year", "must be provided")
	v.Check(instrument.ManufactureYear >= 0, "manufacture_year", "must be greater then 0")
	v.Check(instrument.ManufactureYear <= int32(time.Now().Year()), "manufacture_year", "must not be in the future")

	v.Check(instrument.Type != "", "type", "must not be empty")
	v.Check(validator.PermittedValue(instrument.Type, "synthesizer", "guitar"), "type", "must be synthesizer or guitar")

	v.Check(instrument.EstimatedValue != 0, "estimated_value", "must not be empty")
	v.Check(instrument.EstimatedValue >= 0, "estimated_value", "must be positive")

	v.Check(instrument.Condition != "", "condition", "must not be empty")

	v.Check(validator.Unique(instrument.FamousOwners), "famous_owners", "must be unique")
	v.Check(instrument.OwnerUserId != 0, "owner_user_id", "must not be empty")
	v.Check(instrument.OwnerUserId >= 0, "owner_user_id", "must not be positive")
}

// InstrumentModel represents the database layer and provides functionality to interact with the database.
type InstrumentModel struct {
	DB *sql.DB
}

// Insert creates a new instrument in the database.
func (i InstrumentModel) Insert(instrument *Instrument) error {

	query := `
		INSERT INTO instruments (name, manufacturer, manufacture_year, type, estimated_value, condition, description, famous_owners, owner_user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, version`

	args := []any{
		instrument.Name,
		instrument.Manufacturer,
		instrument.ManufactureYear,
		instrument.Type,
		instrument.EstimatedValue,
		instrument.Condition,
		instrument.Description,
		pq.Array(instrument.FamousOwners),
		instrument.OwnerUserId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return i.DB.QueryRowContext(ctx, query, args...).Scan(&instrument.ID, &instrument.CreatedAt, &instrument.Version)
}

// Get retrieves an instrument from the database based on the provided id value.
// Returns ErrRecordNotFound if no data found during retrieve.
func (i InstrumentModel) Get(id int64) (*Instrument, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, name, manufacturer, manufacture_year, type,
			estimated_value, condition, description, famous_owners, owner_user_id, is_swapped, version
			FROM instruments
				WHERE id = $1
					AND is_deleted = FALSE`

	var instrument Instrument

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, id).Scan(
		&instrument.ID,
		&instrument.CreatedAt,
		&instrument.Name,
		&instrument.Manufacturer,
		&instrument.ManufactureYear,
		&instrument.Type,
		&instrument.EstimatedValue,
		&instrument.Condition,
		&instrument.Description,
		pq.Array(&instrument.FamousOwners),
		&instrument.OwnerUserId,
		&instrument.IsSwapped,
		&instrument.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &instrument, nil
}

// GetAll returns all instrumets stored in the database.
func (i InstrumentModel) GetAll(name string, manufacturer string, iType string, famousOwners []string, ownerUserID int64, filters Filters) ([]*Instrument, MetaData, error) {

	query := fmt.Sprintf(`
		SELECT count(*) over(), id, name, manufacturer, manufacture_year, type, estimated_value,
			condition, description, famous_owners, owner_user_id, is_swapped, version
		FROM instruments
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		  AND (lower(manufacturer) = lower($2) OR $2 = '')
			AND (lower(type) = lower($3) OR $3 = '')
			AND (famous_owners @> $4 OR $4 = '{}')
			AND (owner_user_id = $5 OR $5 = 0)
		  AND is_deleted = FALSE
		ORDER BY %s %s, id ASC
		LIMIT $6 OFFSET $7`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name, manufacturer, iType, pq.Array(famousOwners), ownerUserID, filters.limit(), filters.offset(),
	}

	rows, err := i.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, MetaData{}, err
	}

	defer rows.Close()

	totalRecords := 0
	instruments := []*Instrument{}

	for rows.Next() {

		var instrument Instrument

		err := rows.Scan(
			&totalRecords,
			&instrument.ID,
			&instrument.Name,
			&instrument.Manufacturer,
			&instrument.ManufactureYear,
			&instrument.Type,
			&instrument.EstimatedValue,
			&instrument.Condition,
			&instrument.Description,
			pq.Array(&instrument.FamousOwners),
			&instrument.OwnerUserId,
			&instrument.IsSwapped,
			&instrument.Version,
		)

		if err != nil {
			return nil, MetaData{}, err
		}

		instruments = append(instruments, &instrument)
	}

	if err = rows.Err(); err != nil {
		return nil, MetaData{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return instruments, metadata, nil
}

// Update updates the matching instrument in the database with the provided field values.
// Returns ErrRecordnotFound if no target data found during update.
// Returns ErrEditConflict if there was a race condidion during update.
func (i InstrumentModel) Update(instrument *Instrument) error {

	query := `
		UPDATE instruments
			SET name = $1,
			    manufacturer = $2,
					manufacture_year = $3,
					type = $4,
					estimated_value = $5,
					condition = $6,
					description = $7,
					famous_owners = $8,
					owner_user_id = $9,
					is_swapped = $10,
					version = version + 1
		WHERE id = $11
			AND version = $12
		  AND is_deleted = FALSE
		RETURNING version`

	args := []any{
		instrument.Name,
		instrument.Manufacturer,
		instrument.ManufactureYear,
		instrument.Type,
		instrument.EstimatedValue,
		instrument.Condition,
		instrument.Description,
		pq.Array(instrument.FamousOwners),
		instrument.OwnerUserId,
		instrument.IsSwapped,
		instrument.ID,
		instrument.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, args...).Scan(&instrument.Version)
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

// Delete deletes the corresponding instrument record with the provided id in the database.
// Returns ErrRecordnotFound if no target data found to delete.
// Returns ErrConflict if the deleted instrument is swapped.
func (i InstrumentModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		SELECT is_swapped FROM instruments
			WHERE id = $1
			  AND is_deleted = FALSE`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var isSwapped bool
	err := i.DB.QueryRowContext(ctx, query, id).Scan(&isSwapped)
	if err != nil {
		return err
	}

	if isSwapped {
		return ErrConflict
	}

	query = `
		UPDATE instruments
			SET is_deleted = TRUE, deleted_at = NOW()
		WHERE ID = $1
		  AND is_deleted = FALSE`

	result, err := i.DB.ExecContext(ctx, query, id)
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
