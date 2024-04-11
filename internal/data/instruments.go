package data

import "time"

type Instrument struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"-"`
	IsDeleted       bool      `json:"-"`
	DeletedAt       time.Time `json:"-"`
	Name            string    `json:"name"`
	Manufacturer    string    `json:"manufacturer"`
	ManufactureYear string    `json:"manufacture_year"`
	Type            string    `json:"type"`
	EstimatedValue  float64   `json:"estimated_value"`
	Condition       string    `json:"condition"`
	FamousOwners    []string  `json:"famous_owners"`
}
