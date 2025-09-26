package httpserver

import (
	"context"
	usernoterepo "notes/internal/gateway/userNoteRepo"

	"github.com/stretchr/testify/mock"
)

type mockNoteRepo struct {
	mock.Mock
}

func (m *mockNoteRepo) Get(ctx context.Context, id, accountID int) (usernoterepo.UserNote, error) {
	args := m.Called(ctx, id, accountID)
	return args.Get(0).(usernoterepo.UserNote), args.Error(1)
}

func (m *mockNoteRepo) List(ctx context.Context, accountID int) ([]usernoterepo.UserNote, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]usernoterepo.UserNote), args.Error(1)
}

func (m *mockNoteRepo) Update(ctx context.Context, id, accountID int, title, body string) (bool, error) {
	args := m.Called(ctx, id, accountID, title, body)
	return args.Get(0).(bool), args.Error(1)
}

func (m *mockNoteRepo) Delete(ctx context.Context, id, accountID int) (bool, error) {
	args := m.Called(ctx, id, accountID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *mockNoteRepo) Create(ctx context.Context, accountID int, title, body string) (int, error) {
	args := m.Called(ctx, accountID, accountID)
	return args.Get(0).(int), args.Error(1)
}
