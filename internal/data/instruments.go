package data

import (
	"database/sql"
	"errors"
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

	v.Check(validator.Unique(instrument.FamousOwners), "famous_owners", "must be uniwue")
}

// InstrumentModel represents the database layer and provides functionality to interact with the database.
type InstrumentModel struct {
	DB *sql.DB
}

// Insert creates a new instrument in the database.
func (i InstrumentModel) Insert(instrument *Instrument) error {

	query := `
		INSERT INTO instruments (name, manufacturer, manufacture_year, type, estimated_value, condition, description, famous_owners)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
	}

	return i.DB.QueryRow(query, args...).Scan(&instrument.ID, &instrument.CreatedAt, &instrument.Version)
}

// Get retrieves an instrument from the database based on the provided id value.
// Returns ErrRecordnotFound if no data found to retrieve
func (i InstrumentModel) Get(id int64) (*Instrument, error) {

	if id < 1 {
		return nil, ErrRecordnotFound
	}

	query := `
		SELECT id, created_at, name, manufacturer, manufacture_year, type,
			estimated_value, condition, description, famous_owners, version
			FROM instruments
				WHERE id = $1
					AND is_deleted = FALSE`

	var instrument Instrument

	err := i.DB.QueryRow(query, id).Scan(
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
		&instrument.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordnotFound
		default:
			return nil, err
		}
	}

	return &instrument, nil
}

// GetAll returns all instrumets stored in the database.
func (i InstrumentModel) GetAll() ([]*Instrument, error) {
	return nil, nil
}

// Update updates the matching instrument in the database with the provided field values.
// Returns ErrRecordnotFound if no data found to update
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
					famous_owners = $8
		WHERE id = $9
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
		instrument.ID,
	}

	return i.DB.QueryRow(query, args...).Scan(&instrument.Version)
}

// Delete deletes the corresponding instrument record with the provided id in the database.
// Returns ErrRecordnotFound if no data found to delete
func (i InstrumentModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordnotFound
	}

	query := `
		UPDATE instruments
			SET is_deleted = TRUE, deleted_at = NOW()
		WHERE ID = $1
		  AND is_deleted = FALSE`

	result, err := i.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordnotFound
	}

	return nil

}
