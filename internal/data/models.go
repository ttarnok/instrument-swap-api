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

// Model wraps all database models used in the application.
type Models struct {
	Instruments    InstrumentModel
	Users          UserModel
	Swaps          SwapModel
	StatefulTokens StatefulTokenModel
}

// NewModel rerturn a newly created model based on the specified database connection.
func NewModel(db *sql.DB) Models {
	return Models{
		Instruments:    InstrumentModel{DB: db},
		Users:          UserModel{DB: db},
		Swaps:          SwapModel{DB: db},
		StatefulTokens: StatefulTokenModel{DB: db},
	}
}
