package mocks

import "github.com/ttarnok/instrument-swap-api/internal/data"

// UserModelEmptyMock is a mock implementation for an UserModeler interface.
// Don't store any real users.
type UserModelEmptyMock struct {
}

// NewUserModelEmptyMock creates a new UserModelEmptyMock.
func NewUserModelEmptyMock() *UserModelEmptyMock {
	return &UserModelEmptyMock{}
}

// Insert mocks the instertion of a new user into the model.
// Does not provide any real functionality.
func (u *UserModelEmptyMock) Insert(user *data.User) error {
	return nil
}

// GetAll mocks the retrieval all of the users from the model.
// Does not provide any real functionality.
func (u *UserModelEmptyMock) GetAll() (users []*data.User, err error) {
	return []*data.User{}, nil
}

// GetByEmail mocks the retrieval of a user from the model based on user email.
// Does not provide any real functionality.
func (u *UserModelEmptyMock) GetByEmail(email string) (*data.User, error) {
	return &data.User{}, nil
}

// GetByID mocks the retrieval of a user from the model based on user ID.
// Does not provide any real functionality.
// Returns data.ErrRecordNotFound error if the ID is greater than 99.
func (u *UserModelEmptyMock) GetByID(id int64) (*data.User, error) {
	if id > 99 {
		return nil, data.ErrRecordNotFound
	}
	return &data.User{}, nil
}

// Update mocks the update of a user from the model.
// Does not provide any real functionality.
func (u *UserModelEmptyMock) Update(user *data.User) error {
	return nil
}

// Delete mocks the deletion of a user from the model.
// Does not provide any real functionality.
func (u *UserModelEmptyMock) Delete(id int64) error {
	return nil
}

// GetForStatefulToken mocks the retieval of a user from the model based on its token.
// Does not provide any real functionality.
func (u *UserModelEmptyMock) GetForStatefulToken(tokenScope, tokenPlaintext string) (*data.User, error) {
	return &data.User{}, nil
}
