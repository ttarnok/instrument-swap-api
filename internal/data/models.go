package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordnotFound represents the miss of the requested data.
	ErrRecordnotFound = errors.New("record not found")
	// ErrEditConflict reprezents a race condition error during updates
	ErrEditConflict = errors.New("edit conflict")
)

// Model wraps all database models used in the application.
type Models struct {
	Instruments    InstrumentModel
	Users          UserModel
	StatefulTokens StatefulTokenModel
}

// NewModel rerturn a newly created model based on the specified database connection.
func NewModel(db *sql.DB) Models {
	return Models{
		Instruments: InstrumentModel{DB: db},
		Users:       UserModel{DB: db},
	}
}
