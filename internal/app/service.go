package app

import (
	"context"
	"users/internal/app/user"
	"users/internal/domain"
	"users/pkg/logger"
)

type UserServiceCommands interface {
	user.UserCommands
}
type UserServiceQueries interface {
	user.UserQueries
}

// NewUserServiceQueries creates an instance of User Queries that satisfies UserServiceQueries interface
func NewUserServiceQueries(logger logger.Interface, queries domain.UserRepoQueries) UserServiceQueries {
	return user.NewUserUseCaseQueries(logger, queries)
}

// NewUserServiceCommands creates an instance of User Commands that satisfies UserServiceCommands interface
func NewUserServiceCommands(logger logger.Interface, transaction domain.Transaction, commands domain.UserRepoCommands, outboxCommands domain.OutboxRepoCommands) UserServiceCommands {
	return user.NewUserUseCaseCommands(logger, commands, transaction, outboxCommands)
}

// HealthCheckQueries is an interface for checking the health of application dependencies
type HealthCheckQueries interface {
	Check(ctx context.Context) bool
}
