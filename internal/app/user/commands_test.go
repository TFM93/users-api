package user

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	domainMocks "users/gen/mocks/users/domain"
	loggerMocks "users/gen/mocks/users/pkg/logger"
	"users/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func Test_validatePassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "ok",
			args:    args{password: "Password1!"},
			wantErr: nil,
		},
		{
			name:    "not ok length",
			args:    args{password: "P1234"},
			wantErr: domain.ErrInvalidPW,
		},
		{
			name:    "not ok without numbers",
			args:    args{password: "Password!"},
			wantErr: domain.ErrInvalidPW,
		},
		{
			name:    "not ok without letters",
			args:    args{password: "1234567"},
			wantErr: domain.ErrInvalidPW,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.args.password)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
		})
	}
}

func Test_hashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "pw bigger than 72 bytes",
			args: args{
				password: "VeryLongPasswordThatExceedsSeventyTwoCharactersInLengthWhichIsDefinitelyLongEnoughToTestThePasswordHashFunction@123456789",
			},
			wantErr: fmt.Errorf("bcrypt: password length exceeds 72 bytes"),
		},
		{
			name: "success",
			args: args{
				password: "Password1!",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashPassword(tt.args.password)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(got), []byte(tt.args.password)))

		})
	}
}

func Test_userUseCaseCommands_CreateUser(t *testing.T) {
	mockedLogger := loggerMocks.NewInterface(t)
	repoCommandsMock := domainMocks.NewUserRepoCommands(t)
	outboxCommandsMock := domainMocks.NewOutboxRepoCommands(t)
	transactionMock := domainMocks.NewTransaction(t)
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"
	mockedLogger.On("Debug", mock.Anything, mock.Anything)

	exampleAddUserReq := AddUserRequest{
		FirstName:      "first",
		LastName:       "last",
		NickName:       "nick",
		CountryISOCode: "UK",
		Email:          "email@email.pt",
		Password:       "Password1!",
	}
	addUserReqMap, err := json.Marshal(exampleAddUserReq)
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
		req AddUserRequest
	}
	tests := []struct {
		name          string
		args          args
		want          string
		expectedMocks func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands)
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: exampleAddUserReq,
			},
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(nil).Once()
				commands.On("SaveUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Password != "" &&
						u.FirstName == "first" &&
						u.Email == "email@email.pt" &&
						u.LastName == "last" &&
						u.NickName == "nick"
				})).Return(expectedUserID, nil).Once()
				outbox.On("AddEvent", mock.Anything,
					&domain.Event{Type: "CreateUser", Payload: addUserReqMap}).Return("0d913f6a-497b-4305-b3d1-3f53657e3a27", nil).Once()
			},
			want:    expectedUserID,
			wantErr: nil,
		},
		{
			name: "failed to save user",
			args: args{
				ctx: context.Background(),
				req: AddUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "email@email.pt",
					Password:       "Password1!",
				},
			},
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(domain.ErrInternal).Once()
				commands.On("SaveUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Password != "" &&
						u.FirstName == "first" &&
						u.Email == "email@email.pt" &&
						u.LastName == "last" &&
						u.NickName == "nick"
				})).Return("", domain.ErrInternal).Once()
				l.On("Warn", mock.Anything, domain.ErrInternal).Return().Once()
			},
			want:    "",
			wantErr: domain.ErrInternal,
		},
		{
			name: "failed to add event to outbox",
			args: args{
				ctx: context.Background(),
				req: exampleAddUserReq,
			},
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(domain.ErrInternal).Once()
				commands.On("SaveUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Password != "" &&
						u.FirstName == "first" &&
						u.Email == "email@email.pt" &&
						u.LastName == "last" &&
						u.NickName == "nick"
				})).Return(expectedUserID, nil).Once()
				outbox.On("AddEvent", mock.Anything, &domain.Event{Type: "CreateUser", Payload: addUserReqMap}).Return("", domain.ErrInternal).Once()
				l.On("Warn", mock.Anything, domain.ErrInternal).Return().Once()
			},
			want:    "",
			wantErr: domain.ErrInternal,
		}, {
			name: "invalid pw",
			args: args{
				ctx: context.Background(),
				req: AddUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "email@email.pt",
					Password:       "P",
				},
			},
			want:    "",
			wantErr: domain.ErrInvalidPW,
		},
		{
			name: "invalid hash",
			args: args{
				ctx: context.Background(),
				req: AddUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "email@email.pt",
					Password:       "VeryLongPasswordThatExceedsSeventyTwoCharactersInLengthWhichIsDefinitelyLongEnoughToTestThePasswordHashFunction@123456789",
				},
			},
			want:    "",
			wantErr: domain.ErrInvalidPW,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := NewUserUseCaseCommands(mockedLogger, repoCommandsMock, transactionMock, outboxCommandsMock)
			if tt.expectedMocks != nil {
				tt.expectedMocks(mockedLogger, repoCommandsMock, transactionMock, outboxCommandsMock)
			}
			got, err := commands.CreateUser(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if got != tt.want {
				t.Errorf("userUseCaseCommands.CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userUseCaseCommands_DeleteUser(t *testing.T) {
	mockedLogger := loggerMocks.NewInterface(t)
	repoCommandsMock := domainMocks.NewUserRepoCommands(t)
	outboxCommandsMock := domainMocks.NewOutboxRepoCommands(t)
	transactionMock := domainMocks.NewTransaction(t)
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"
	mockedLogger.On("Warn", mock.Anything, mock.Anything)

	delUserReqMap, err := json.Marshal(map[string]interface{}{
		"ID": expectedUserID,
	})
	assert.NoError(t, err)

	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name          string
		expectedMocks func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands)
		args          args
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: expectedUserID,
			},
			wantErr: nil,
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(nil).Once()
				commands.On("DeleteUser", mock.Anything, expectedUserID).Return(nil).Once()
				outbox.On("AddEvent", mock.Anything, &domain.Event{Type: "DeleteUser", Payload: delUserReqMap}).Return("0d913f6a-497b-4305-b3d1-3f53657e3a27", nil).Once()
			},
		}, {
			name: "invalid user id",
			args: args{
				ctx:    context.Background(),
				userID: "expectedUserID",
			},
			wantErr: domain.ErrInvalidUserID,
		}, {
			name: "failed to delete user",
			args: args{
				ctx:    context.Background(),
				userID: expectedUserID,
			},
			wantErr: domain.ErrUserNotFound,
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(domain.ErrUserNotFound).Once()
				commands.On("DeleteUser", mock.Anything, expectedUserID).Return(domain.ErrUserNotFound).Once()
			},
		}, {
			name: "failed to send event",
			args: args{
				ctx:    context.Background(),
				userID: expectedUserID,
			},
			wantErr: domain.ErrInternal,
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(fmt.Errorf("something went wrong")).Once()
				commands.On("DeleteUser", mock.Anything, expectedUserID).Return(nil).Once()
				outbox.On("AddEvent", mock.Anything, &domain.Event{Type: "DeleteUser", Payload: delUserReqMap}).Return("", fmt.Errorf("something went wrong")).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := NewUserUseCaseCommands(mockedLogger, repoCommandsMock, transactionMock, outboxCommandsMock)
			if tt.expectedMocks != nil {
				tt.expectedMocks(mockedLogger, repoCommandsMock, transactionMock, outboxCommandsMock)
			}
			err := commands.DeleteUser(tt.args.ctx, tt.args.userID)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
		})
	}
}

func Test_userUseCaseCommands_UpdateUser(t *testing.T) {
	mockedLogger := loggerMocks.NewInterface(t)
	repoCommandsMock := domainMocks.NewUserRepoCommands(t)
	outboxCommandsMock := domainMocks.NewOutboxRepoCommands(t)
	transactionMock := domainMocks.NewTransaction(t)
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"
	mockedLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything)

	exampleUpdateUserReq := UpdateUserRequest{
		ID:             expectedUserID,
		FirstName:      "first",
		LastName:       "last",
		NickName:       "nick",
		CountryISOCode: "UK",
		Email:          "email@email.pt",
	}
	updateUserReqMap, err := json.Marshal(exampleUpdateUserReq)
	assert.NoError(t, err)

	type args struct {
		ctx context.Context
		req UpdateUserRequest
	}
	tests := []struct {
		name          string
		expectedMocks func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands)
		args          args
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: exampleUpdateUserReq,
			},
			wantErr: nil,
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(nil).Once()
				commands.On("UpdateUser", mock.Anything, mock.Anything).Return(nil).Once()
				outbox.On("AddEvent", mock.Anything, &domain.Event{Type: "UpdateUser", Payload: updateUserReqMap}).Return("0d913f6a-497b-4305-b3d1-3f53657e3a27", nil).Once()
			},
		}, {
			name: "invalid user id",
			args: args{
				ctx: context.Background(),
				req: UpdateUserRequest{ID: "invalid"},
			},
			wantErr: domain.ErrInvalidUserID,
		}, {
			name: "failed to update user",
			args: args{
				ctx: context.Background(),
				req: exampleUpdateUserReq,
			},
			wantErr: domain.ErrUserNotFound,
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(domain.ErrUserNotFound).Once()
				commands.On("UpdateUser", mock.Anything, mock.Anything).Return(domain.ErrUserNotFound).Once()
			},
		}, {
			name: "failed to send event",
			args: args{
				ctx: context.Background(),
				req: exampleUpdateUserReq,
			},
			wantErr: domain.ErrInternal,
			expectedMocks: func(l *loggerMocks.Interface, commands *domainMocks.UserRepoCommands, tr *domainMocks.Transaction, outbox *domainMocks.OutboxRepoCommands) {
				tr.On("BeginTx", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(ctx context.Context) error)
					fn(args.Get(0).(context.Context))
				}).Return(fmt.Errorf("something went wrong")).Once()
				commands.On("UpdateUser", mock.Anything, mock.Anything).Return(nil).Once()
				outbox.On("AddEvent", mock.Anything, &domain.Event{Type: "UpdateUser", Payload: updateUserReqMap}).Return("", fmt.Errorf("something went wrong")).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := NewUserUseCaseCommands(mockedLogger, repoCommandsMock, transactionMock, outboxCommandsMock)
			if tt.expectedMocks != nil {
				tt.expectedMocks(mockedLogger, repoCommandsMock, transactionMock, outboxCommandsMock)
			}
			err := commands.UpdateUser(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
		})
	}
}
