package user

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
	domainMocks "users/gen/mocks/users/domain"
	loggerMocks "users/gen/mocks/users/pkg/logger"
	"users/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_userUseCaseQueries_GetUser(t *testing.T) {
	mockedLogger := loggerMocks.NewInterface(t)
	repoQueriesMock := domainMocks.NewUserRepoQueries(t)
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name          string
		expectedMocks func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries)
		args          args
		wantDu        *domain.User
		wantErr       error
	}{
		{
			name:    "invalid user id",
			args:    args{ctx: context.Background(), userID: "invalid"},
			wantDu:  &domain.User{},
			wantErr: domain.ErrInvalidUserID,
		},
		{
			name:    "failed to fetch users",
			args:    args{ctx: context.Background(), userID: expectedUserID},
			wantDu:  &domain.User{},
			wantErr: domain.ErrInternal,
			expectedMocks: func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries) {
				queries.On("GetUser", mock.Anything, expectedUserID).Return(nil, fmt.Errorf("something went wrong")).Once()
				mockedLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Once()

			},
		},
		{
			name:    "user not found",
			args:    args{ctx: context.Background(), userID: expectedUserID},
			wantDu:  &domain.User{},
			wantErr: domain.ErrUserNotFound,
			expectedMocks: func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries) {
				queries.On("GetUser", mock.Anything, expectedUserID).Return(nil, domain.ErrUserNotFound).Once()
			},
		},
		{
			name:    "success",
			args:    args{ctx: context.Background(), userID: expectedUserID},
			wantDu:  &domain.User{NickName: "nick"},
			wantErr: nil,
			expectedMocks: func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries) {
				queries.On("GetUser", mock.Anything, expectedUserID).Return(&domain.User{NickName: "nick"}, nil).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserUseCaseQueries(mockedLogger, repoQueriesMock)
			if tt.expectedMocks != nil {
				tt.expectedMocks(mockedLogger, repoQueriesMock)
			}
			gotDu, err := uc.GetUser(tt.args.ctx, tt.args.userID)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(gotDu, tt.wantDu) {
				t.Errorf("userUseCaseQueries.GetUser() = %v, want %v", gotDu, tt.wantDu)
			}
		})
	}
}

func Test_userUseCaseQueries_ListUsers(t *testing.T) {
	tnow := time.Now()
	mockedLogger := loggerMocks.NewInterface(t)
	repoQueriesMock := domainMocks.NewUserRepoQueries(t)
	domainUser1 := domain.User{
		ID:             uuid.MustParse("c12e23f3-f5e3-41bc-aeca-9d66bd0b96a3"),
		FirstName:      "nick",
		LastName:       "nock",
		NickName:       "ooo",
		Email:          "sads@sdas.pt",
		CountryISOCode: "PT",
		CreatedAt:      tnow,
		UpdatedAt:      tnow,
	}
	domainUser2 := domain.User{
		ID:             uuid.MustParse("c12e23f3-f5e3-41bc-f5e3-9d66bd0b96a3"),
		FirstName:      "nick2",
		LastName:       "nock2",
		NickName:       "ooo2",
		Email:          "sads2@sdas.pt",
		CountryISOCode: "UK",
		CreatedAt:      tnow,
		UpdatedAt:      tnow,
	}
	expectedCursorTime := time.Date(2024, time.August, 22, 20, 9, 11, 938220000, time.Local)
	var nilt *time.Time
	type args struct {
		ctx context.Context
		req ListUsersRequest
	}
	tests := []struct {
		name          string
		expectedMocks func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries)
		args          args
		wantUsers     []*domain.User
		wantNextCur   bool
		wantErr       error
	}{
		{
			name: "error listing users",
			args: args{
				ctx: context.Background(),
				req: ListUsersRequest{
					Cursor: "",
					Limit:  2,
				},
			},
			expectedMocks: func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries) {
				l.On("Debug", mock.Anything, mock.Anything).Once()
				queries.On("ListUsers", mock.Anything, "", nilt, int32(2), domain.UserSearchFilters{}).Return(
					[]*domain.User{}, fmt.Errorf("something went wrong")).Once()
			},
			wantUsers:   []*domain.User{},
			wantNextCur: false,
			wantErr:     domain.ErrInternal,
		},
		{
			name: "more pages",
			args: args{
				ctx: context.Background(),
				req: ListUsersRequest{
					Cursor: "",
					Limit:  2,
				},
			},
			expectedMocks: func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries) {
				queries.On("ListUsers", mock.Anything, "", nilt, int32(2), domain.UserSearchFilters{}).Return(
					[]*domain.User{&domainUser1, &domainUser2}, nil).Once()
			},
			wantUsers:   []*domain.User{&domainUser1, &domainUser2},
			wantNextCur: true,
			wantErr:     nil,
		},
		{
			name: "with cursor",
			args: args{
				ctx: context.Background(),
				req: ListUsersRequest{
					Cursor: "MjAyNC0wOC0yMlQyMDowOToxMS45MzgyMiswMTowMHxjMTJlMjNmMy1mNWUzLTQxYmMtYWVjYS05ZDY2YmQwYjk2YTM=",
					Limit:  2,
				},
			},
			expectedMocks: func(l *loggerMocks.Interface, queries *domainMocks.UserRepoQueries) {
				queries.On("ListUsers", mock.Anything, "c12e23f3-f5e3-41bc-aeca-9d66bd0b96a3", &expectedCursorTime, int32(2), domain.UserSearchFilters{}).Return(
					[]*domain.User{&domainUser1, &domainUser2}, nil).Once()
			},
			wantUsers:   []*domain.User{&domainUser1, &domainUser2},
			wantNextCur: true,
			wantErr:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewUserUseCaseQueries(mockedLogger, repoQueriesMock)
			if tt.expectedMocks != nil {
				tt.expectedMocks(mockedLogger, repoQueriesMock)
			}
			gotUsers, gotNextCur, err := uc.ListUsers(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(gotUsers, tt.wantUsers) {
				t.Errorf("userUseCaseQueries.ListUsers() gotUsers = %v, want %v", gotUsers, tt.wantUsers)
			}
			if (len(gotNextCur) > 0) != tt.wantNextCur {
				t.Errorf("userUseCaseQueries.ListUsers() gotNextCur = %v, want %v", gotNextCur, tt.wantNextCur)
			}
		})
	}
}
