package mocks

import "github.com/ttarnok/instrument-swap-api/internal/data"

// UserModelEmptyMock is a mock implementation for an UserModeler interface.
// Don't store any users, never returns any error.
type UserModelEmptyMock struct {
}

func NewUserModelEmptyMock() *UserModelEmptyMock {
	return &UserModelEmptyMock{}
}

func (u *UserModelEmptyMock) Insert(user *data.User) error {
	return nil
}
func (u *UserModelEmptyMock) GetAll() (users []*data.User, err error) {
	return []*data.User{}, nil
}

func (u *UserModelEmptyMock) GetByEmail(email string) (*data.User, error) {
	return &data.User{}, nil
}

func (u *UserModelEmptyMock) GetByID(id int64) (*data.User, error) {
	return &data.User{}, nil
}

func (u *UserModelEmptyMock) Update(user *data.User) error {
	return nil
}

func (u *UserModelEmptyMock) Delete(id int64) error {
	return nil
}

func (u *UserModelEmptyMock) GetForStatefulToken(tokenScope, tokenPlaintext string) (*data.User, error) {
	return &data.User{}, nil
}
