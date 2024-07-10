// Package mocks contains mocks for the database.
package mocks

import (
	"slices"
	"sync"

	"github.com/ttarnok/instrument-swap-api/internal/data"
)

// InstrumentModelMock is a mock implementation for an instrument model.
type InstrumentModelMock struct {
	db []*data.Instrument
	sync.Mutex
}

// NewEmptyInstrumentModelMock returns a new InstrumentModelMock with empty internal store.
func NewEmptyInstrumentModelMock() *InstrumentModelMock {
	return &InstrumentModelMock{}
}

// NewNonEmptyInstrumentModelMock returns a new InstrumentModelMock filled with the input data.
func NewNonEmptyInstrumentModelMock(input []*data.Instrument) *InstrumentModelMock {
	return &InstrumentModelMock{
		db: input,
	}
}

// Insert inserts an instrument into the mocked database.
func (im *InstrumentModelMock) Insert(instrument *data.Instrument) error {
	im.Lock()
	defer im.Unlock()
	im.db = append(im.db, instrument)
	return nil

}

// Get retrieves an instrument from the mocked database.
func (im *InstrumentModelMock) Get(id int64) (*data.Instrument, error) {
	for _, i := range im.db {
		if i.ID == id {
			return i, nil
		}
	}

	return nil, data.ErrRecordNotFound
}

// GetAll returns all instruments from the mocked database.
func (im *InstrumentModelMock) GetAll(name string, manufacturer string, iType string, famousOwners []string, ownerUserID int64, filters data.Filters) (instruments []*data.Instrument, metaData data.MetaData, err error) {
	return im.db, data.MetaData{}, nil
}

// Update updates an instrument record in the mocked database.
func (im *InstrumentModelMock) Update(instrument *data.Instrument) error {
	im.Lock()
	defer im.Unlock()
	for index, i := range im.db {
		if i.ID == instrument.ID {
			im.db[index] = instrument
			return nil
		}
	}
	return data.ErrEditConflict
}

// Delete deletes an instrument from the mocked database.
// If the provided id is not found, returns data.ErrRecordNotFound.
// If the provided id is 999, returns data.ErrConflict.
func (im *InstrumentModelMock) Delete(id int64) error {

	if id == 999 {
		return data.ErrConflict
	}

	im.Lock()
	defer im.Unlock()

	indexToDel := -1
	for index, i := range im.db {
		if i != nil && i.ID == id {
			indexToDel = index
			break
		}
	}
	if indexToDel == -1 {
		return data.ErrRecordNotFound
	}
	im.db = slices.Delete(im.db, indexToDel, indexToDel+1)
	return nil
}
