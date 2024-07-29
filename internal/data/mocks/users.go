package mocks

import (
	"github.com/ttarnok/instrument-swap-api/internal/data"
)

// UserModelMock is a mock implementation for an UserModeler interface.
type UserModelMock struct {
	users []*data.User
}

// NewEmptyUserModelMock creates a new empty UserModelMock.
func NewEmptyUserModelMock() *UserModelMock {
	return &UserModelMock{users: nil}
}

// NewUserModelMock returns a new UserModelMock based on the input users.
func NewUserModelMock(users []*data.User) *UserModelMock {
	return &UserModelMock{users: users}
}

// Insert mocks the instertion of a new user into the model.
// Does not provide any real functionality.
func (u *UserModelMock) Insert(user *data.User) error {
	u.users = append(u.users, user)
	return nil
}

// GetAll mocks the retrieval all of the users from the model.
func (u *UserModelMock) GetAll() (users []*data.User, err error) {
	return u.users, nil
}

// GetByEmail mocks the retrieval of a user from the model based on user email.
// Returns data.ErrRecordNotFound error if a user with the given email is not stored.
func (u *UserModelMock) GetByEmail(email string) (*data.User, error) {
	for _, u := range u.users {
		if u.Email == email {
			return u, nil
		}
	}

	return &data.User{}, data.ErrRecordNotFound
}

// GetByID mocks the retrieval of a user from the model based on user ID.
// Returns data.ErrRecordNotFound error if a user with the given ID is not stored.
// If the data store is empty and the provided user id is greater than 99, returns data.ErrRecordNotFound.
func (u *UserModelMock) GetByID(id int64) (*data.User, error) {
	if u.users == nil {
		if id > 99 {
			return nil, data.ErrRecordNotFound
		}
		return &data.User{ID: id}, nil
	}

	for _, u := range u.users {
		if u.ID == id {
			return u, nil
		}
	}

	return nil, data.ErrRecordNotFound
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
