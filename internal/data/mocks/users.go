package mocks

import (
	"slices"
	"sync"

	"github.com/ttarnok/instrument-swap-api/internal/data"
)

// UserModelMock is a mock implementation for an UserModeler interface.
type UserModelMock struct {
	users []*data.User
	sync.Mutex
}

// NewEmptyUserModelMock creates a new empty UserModelMock.
func NewEmptyUserModelMock() *UserModelMock {
	return &UserModelMock{users: nil}
}

// NewUserModelMock returns a new UserModelMock based on the input users.
func NewUserModelMock(users []*data.User) *UserModelMock {
	uc := make([]*data.User, len(users))
	copy(uc, users)
	return &UserModelMock{users: uc}
}

// Insert mocks the instertion of a new user into the model.
// Does not provide any real functionality.
func (u *UserModelMock) Insert(user *data.User) error {
	u.Lock()
	defer u.Unlock()

	for _, u := range u.users {
		if u.Email == user.Email {
			return data.ErrDuplicateEmail
		}
	}
	u.users = append(u.users, user)
	return nil
}

// GetAll mocks the retrieval all of the users from the model.
func (u *UserModelMock) GetAll() (users []*data.User, err error) {
	u.Lock()
	defer u.Unlock()

	return u.users, nil
}

// GetByEmail mocks the retrieval of a user from the model based on user email.
// Returns data.ErrRecordNotFound error if a user with the given email is not stored.
func (u *UserModelMock) GetByEmail(email string) (*data.User, error) {
	u.Lock()
	defer u.Unlock()

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
	u.Lock()
	defer u.Unlock()

	if u.users == nil {
		if id > 99 {
			return nil, data.ErrRecordNotFound
		}

		return &data.User{ID: id}, nil
	}

	for _, uRec := range u.users {
		if uRec.ID == id {
			return uRec, nil
		}
	}
	return nil, data.ErrRecordNotFound
}

// Update mocks the update of a user from the model.
func (u *UserModelMock) Update(user *data.User) error {
	u.Lock()
	defer u.Unlock()

	for i := range u.users {
		if u.users[i].ID == user.ID {
			u.users[i] = user
			return nil
		}
	}

	return data.ErrRecordNotFound
}

// Delete mocks the deletion of a user from the model.
// If the given id is not found, returns data.ErrRecordNotFound.
func (u *UserModelMock) Delete(id int64) error {
	u.Lock()
	defer u.Unlock()

	for i, user := range u.users {
		if user.ID == id {
			u.users = slices.Delete(u.users, i, i+1)
			return nil
		}
	}
	return data.ErrRecordNotFound
}

// GetForStatefulToken mocks the retieval of a user from the model based on its token.
// Does not provide any real functionality.
func (u *UserModelMock) GetForStatefulToken(tokenScope, tokenPlaintext string) (*data.User, error) {
	u.Lock()
	defer u.Unlock()

	return &data.User{}, nil
}
