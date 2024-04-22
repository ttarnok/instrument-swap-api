package data

import (
	"time"

	"github.com/ttarnok/instrument-swap-api/internal/validator"
)

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
	FamousOwners    []string  `json:"famous_owners"`
}

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
