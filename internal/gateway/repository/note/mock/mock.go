package notemock

import (
	"context"
	noterepo "notes/internal/gateway/repository/note"

	"github.com/stretchr/testify/mock"
)

type MockNoteRepo struct {
	mock.Mock
}

func (m *MockNoteRepo) Get(ctx context.Context, id, accountID int) (noterepo.Note, error) {
	args := m.Called(ctx, id, accountID)
	return args.Get(0).(noterepo.Note), args.Error(1)
}

func (m *MockNoteRepo) List(ctx context.Context, accountID int) ([]noterepo.Note, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]noterepo.Note), args.Error(1)
}

func (m *MockNoteRepo) Update(ctx context.Context, id, accountID int, title, body string) (bool, error) {
	args := m.Called(ctx, id, accountID, title, body)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockNoteRepo) Delete(ctx context.Context, id, accountID int) (bool, error) {
	args := m.Called(ctx, id, accountID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockNoteRepo) Create(ctx context.Context, accountID int, title, body string) (int, error) {
	args := m.Called(ctx, accountID, accountID)
	return args.Get(0).(int), args.Error(1)
}
