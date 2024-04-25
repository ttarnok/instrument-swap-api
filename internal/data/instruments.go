package data

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

// Instrument represents an instrument record in the apprication.
type Instrument struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"-"`
	IsDeleted       bool      `json:"-"`
	DeletedAt       time.Time `json:"-"`
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
func (i InstrumentModel) Get(id int64) (*Instrument, error) {
	return nil, nil
}

// GetAll returns all instrumets stored in the database.
func (i InstrumentModel) GetAll() ([]*Instrument, error) {
	return nil, nil
}

// Update updates the matching instrument in the database with the provided field values.
func (i InstrumentModel) Update(instrument *Instrument) error {
	return nil
}

// Delete deletes the corresponding instrument record with the provided id in the database.
func (i InstrumentModel) Delete(id int64) error {
	return nil
}
