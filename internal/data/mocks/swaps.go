package mocks

import (
	"errors"

	"github.com/ttarnok/instrument-swap-api/internal/data"
)

// SwapModelMock is a mock implementation for an instrument model.
type SwapModelMock struct {
	db []*data.Swap
}

// NewSwapModelMock returns a new SwapModelMock based on the given db slice.
func NewSwapModelMock(db []*data.Swap) *SwapModelMock {
	return &SwapModelMock{db: db}
}

// GetAllForUser is a mocked method for SwapModelMock.
// Returns all swaps stored in the struct.
func (s *SwapModelMock) GetAllForUser(userID int64) ([]*data.Swap, error) {
	if s.db == nil {
		return nil, errors.New("error")
	}
	return s.db, nil
}

// Get is a mocked mothof for SwapModelMock.
// Returns the stored swap with the given id, returns an error otherwise.
func (s *SwapModelMock) Get(id int64) (*data.Swap, error) {
	for _, swap := range s.db {
		if swap.ID == id {
			return swap, nil
		}
	}
	return nil, data.ErrRecordNotFound
}

// GetByInstrumentID is a mocked mothof for SwapModelMock.
func (s *SwapModelMock) GetByInstrumentID(id int64) (*data.Swap, error) {
	return &data.Swap{}, nil
}

// Insert is a mocked mothof for SwapModelMock.
func (s *SwapModelMock) Insert(swap *data.Swap) error {
	return nil
}

// Update is a mocked mothof for SwapModelMock.
func (s *SwapModelMock) Update(swap *data.Swap) error {
	return nil
}
