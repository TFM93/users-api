package postgresql

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	loggermocks "users/gen/mocks/users/pkg/logger"
	dbmocks "users/gen/mocks/users/pkg/postgresql"
	"users/internal/domain"
	"users/pkg/postgresql"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest)
	return args.Error(0)
}

func TestUserCommandsRepo_SaveUser(t *testing.T) {
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"

	mockDB := dbmocks.NewInterface(t)
	mockLogger := loggermocks.NewInterface(t)
	mockDBProvider := dbmocks.NewDBProvider(t)
	mockRow := new(MockRow)
	r := NewUserCommandsRepo(mockDB, mockLogger)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()
	mockDB.On("GetPool").Return(mockDBProvider)

	type args struct {
		ctx  context.Context
		user *domain.User
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func()
		wantId        string
		wantErr       error
	}{
		{
			name: "user saved",
			args: args{
				ctx: context.Background(),
				user: &domain.User{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "first@test.pt",
					Password:       "someHashHere",
				},
			},
			expectedMocks: func() {
				mockDBProvider.On("QueryRow", mock.Anything,
					"INSERT INTO users (first_name, last_name, country_iso_code, nickname, email, pw) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
					"first", "last", "UK", "nick", "first@test.pt", "someHashHere").Return(mockRow).Once()
				mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
					// change the value of the scan argument
					arg := args.Get(0).([]interface{})
					*arg[0].(*string) = expectedUserID
				}).Return(nil).Once()
			},
			wantId:  expectedUserID,
			wantErr: nil,
		},
		{
			name: "user already exists",
			args: args{
				ctx: context.Background(),
				user: &domain.User{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "first@test.pt",
					Password:       "someHashHere",
				},
			},
			expectedMocks: func() {
				mockDBProvider.On("QueryRow", mock.Anything,
					"INSERT INTO users (first_name, last_name, country_iso_code, nickname, email, pw) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
					"first", "last", "UK", "nick", "first@test.pt", "someHashHere").Return(mockRow).Once()
				mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
					// change the value of the scan argument
					arg := args.Get(0).([]interface{})
					*arg[0].(*string) = expectedUserID
				}).Return(&pgconn.PgError{Code: "23505"}).Once() // NOTE: pgconn should not be directly accessed here, kept it only for this hiring challenge
			},
			wantId:  "",
			wantErr: domain.ErrUserAlreadyExists,
		},
		{
			name: "create user with another error",
			args: args{
				ctx: context.Background(),
				user: &domain.User{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "first@test.pt",
					Password:       "someHashHere",
				},
			},
			expectedMocks: func() {
				mockDBProvider.On("QueryRow", mock.Anything,
					"INSERT INTO users (first_name, last_name, country_iso_code, nickname, email, pw) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
					"first", "last", "UK", "nick", "first@test.pt", "someHashHere").Return(mockRow).Once()
				mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
					// change the value of the scan argument
					arg := args.Get(0).([]interface{})
					*arg[0].(*string) = expectedUserID
				}).Return(fmt.Errorf("something went wrong")).Once()
			},
			wantId:  "",
			wantErr: domain.ErrInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks()
			}
			gotId, err := r.SaveUser(tt.args.ctx, tt.args.user)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if gotId != tt.wantId {
				t.Errorf("UserCommandsRepo.SaveUser() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func TestUserCommandsRepo_db(t *testing.T) {

	mockDB := dbmocks.NewInterface(t)
	mockLogger := loggermocks.NewInterface(t)
	mockTx := dbmocks.NewTx(t)
	mockDBProvider := dbmocks.NewDBProvider(t)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func()
		want          postgresql.DBProvider
	}{
		{
			name: "return tx",
			args: args{
				ctx: context.WithValue(context.Background(), domain.TxKey, mockTx),
			},
			want: mockTx,
		},
		{
			name: "return from pool",
			args: args{
				ctx: context.Background(),
			},
			expectedMocks: func() {
				mockDB.On("GetPool").Return(mockDBProvider).Once()
			},
			want: mockDBProvider,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &userCommandsRepo{
				pg: mockDB,
				l:  mockLogger,
			}
			if tt.expectedMocks != nil {
				tt.expectedMocks()
			}
			if got := r.db(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserCommandsRepo.db() = %v, want %v", got, tt.want)
			}
		})
	}
}
