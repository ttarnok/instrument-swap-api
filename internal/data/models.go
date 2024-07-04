// Package data contains the database access layer of the application.
package data

import (
	"database/sql"
	"errors"
)

// Generic data related errors.
// These errors can be tested using errors.Is.
var (
	ErrRecordNotFound = errors.New("record not found") // "record not found"
	ErrEditConflict   = errors.New("edit conflict")    // "edit conflict"
	ErrConflict       = errors.New("conflict")         // "conflict"
)

// InstrumentModeler interface abstracts a model for instruments.
type InstrumentModeler interface {
	Insert(instrument *Instrument) error
	Get(id int64) (*Instrument, error)
	GetAll(name string, manufacturer string, iType string, famousOwners []string, ownerUserID int64, filters Filters) (instruments []*Instrument, metaData MetaData, err error)
	Update(instrument *Instrument) error
	Delete(id int64) error
}

// Models wraps all database models used in the application.
type Models struct {
	Instruments    InstrumentModeler
	Users          UserModel
	Swaps          SwapModel
	StatefulTokens StatefulTokenModel
}

// NewModel rerturn a newly created model based on the specified database connection.
func NewModel(db *sql.DB) Models {
	return Models{
		Instruments:    &InstrumentModel{DB: db},
		Users:          UserModel{DB: db},
		Swaps:          SwapModel{DB: db},
		StatefulTokens: StatefulTokenModel{DB: db},
	}
}
