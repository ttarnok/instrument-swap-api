package mocks

import (
	"sync"
)

// BlacklistServiceMock represents a Mock for BlacklistService.
type BlacklistServiceMock struct {
	store map[string]string
	sync.Mutex
}

// NewBlacklistServiceMock creates a new empty *BlacklistServiceMock.
func NewBlacklistServiceMock() *BlacklistServiceMock {
	return &BlacklistServiceMock{
		store: make(map[string]string),
	}
}

// NewBlacklistServiceMockWithData creates a new BlacklistServiceMock, filled with the given token ID values.
func NewBlacklistServiceMockWithData(data []string) *BlacklistServiceMock {
	store := make(map[string]string, len(data))
	for _, tokenID := range data {
		store[tokenID] = "blacklisted"
	}
	return &BlacklistServiceMock{
		store: store,
	}
}

// BlacklistToken blacklists the given token ID.
func (b *BlacklistServiceMock) BlacklistToken(token string) error {
	b.Lock()
	defer b.Unlock()
	b.store[token] = "blacklisted"
	return nil
}

// IsTokenBlacklisted checks wheather the given token ID is blacklisted.
func (b *BlacklistServiceMock) IsTokenBlacklisted(token string) (bool, error) {
	if _, ok := b.store[token]; !ok {
		return false, nil
	}

	return true, nil
}
