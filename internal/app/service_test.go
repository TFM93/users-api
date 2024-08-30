package app

import (
	"reflect"
	"testing"
	mocks "users/gen/mocks/users/domain"
	loggermocks "users/gen/mocks/users/pkg/logger"
	"users/internal/app/user"
	"users/internal/domain"
	"users/pkg/logger"
)

func TestNewUserServiceCommands(t *testing.T) {
	mockLogger := loggermocks.NewInterface(t)
	transactionMock := mocks.NewTransaction(t)
	commandsMock := mocks.NewUserRepoCommands(t)
	outboxCommandsMock := mocks.NewOutboxRepoCommands(t)
	type args struct {
		logger         logger.Interface
		transaction    domain.Transaction
		commands       domain.UserRepoCommands
		outboxCommands domain.OutboxRepoCommands
	}
	tests := []struct {
		name string
		args args
		want UserServiceCommands
	}{
		{
			name: "success",
			args: args{
				logger:         mockLogger,
				transaction:    transactionMock,
				commands:       commandsMock,
				outboxCommands: outboxCommandsMock,
			},
			want: user.NewUserUseCaseCommands(mockLogger, commandsMock, transactionMock, outboxCommandsMock),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUserServiceCommands(tt.args.logger, tt.args.transaction, tt.args.commands, tt.args.outboxCommands); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserServiceCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUserServiceQueries(t *testing.T) {
	mockLogger := loggermocks.NewInterface(t)
	queriesMock := mocks.NewUserRepoQueries(t)
	type args struct {
		logger  logger.Interface
		queries domain.UserRepoQueries
	}
	tests := []struct {
		name string
		args args
		want UserServiceQueries
	}{
		{
			name: "success",
			args: args{
				logger:  mockLogger,
				queries: queriesMock,
			},
			want: user.NewUserUseCaseQueries(mockLogger, queriesMock),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUserServiceQueries(tt.args.logger, tt.args.queries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserServiceQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
