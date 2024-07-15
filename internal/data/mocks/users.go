package mocks

import (
	"github.com/ttarnok/instrument-swap-api/internal/data"
)

// UserModelMock is a mock implementation for an UserModeler interface.
type UserModelMock struct {
	users map[int64]*data.User
}

// NewEmptyUserModelMock creates a new empty UserModelMock.
func NewEmptyUserModelMock() *UserModelMock {
	return &UserModelMock{users: nil}
}

// NewUserModelMock returns a new UserModelMock based on the input users.
func NewUserModelMock(users map[int64]*data.User) *UserModelMock {
	return &UserModelMock{users: users}
}

// Insert mocks the instertion of a new user into the model.
// Does not provide any real functionality.
func (u *UserModelMock) Insert(user *data.User) error {
	return nil
}

// GetAll mocks the retrieval all of the users from the model.
// Does not provide any real functionality.
func (u *UserModelMock) GetAll() (users []*data.User, err error) {
	return []*data.User{}, nil
}

// GetByEmail mocks the retrieval of a user from the model based on user email.
// Does not provide any real functionality.
func (u *UserModelMock) GetByEmail(email string) (*data.User, error) {
	return &data.User{}, nil
}

// GetByID mocks the retrieval of a user from the model based on user ID.
// Returns data.ErrRecordNotFound error if the ID is not stored.
// If the data store is empty and the provided user id is greater than 99, returns data.ErrRecordNotFound.
func (u *UserModelMock) GetByID(id int64) (*data.User, error) {
	if u.users == nil {
		if id > 99 {
			return nil, data.ErrRecordNotFound
		}
		return &data.User{ID: id}, nil
	}

	user, ok := u.users[id]
	if !ok {
		return nil, data.ErrRecordNotFound
	}
	return user, nil
}

// Update mocks the update of a user from the model.
// Does not provide any real functionality.
func (u *UserModelMock) Update(user *data.User) error {
	return nil
}

// Delete mocks the deletion of a user from the model.
// Does not provide any real functionality.
func (u *UserModelMock) Delete(id int64) error {
	return nil
}

// GetForStatefulToken mocks the retieval of a user from the model based on its token.
// Does not provide any real functionality.
func (u *UserModelMock) GetForStatefulToken(tokenScope, tokenPlaintext string) (*data.User, error) {
	return &data.User{}, nil
}
