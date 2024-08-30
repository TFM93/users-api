package user

import (
	"context"
	"errors"
	"time"
	"users/internal/domain"
	"users/pkg/logger"

	"github.com/google/uuid"
)

type ListUsersRequest struct {
	Cursor  string
	Limit   int32
	Filters domain.UserSearchFilters
}

type UserQueries interface {
	// GetUser retrieves a single User based on his id.
	// It returns domain.ErrInvalidUserID if an invalid user id is provided.
	// It returns domain.ErrUserNotFound if the user does not exist.
	// It returns domain.ErrInternal if it fails to fetch from the repository.
	GetUser(ctx context.Context, userID string) (user *domain.User, err error)

	// ListUsers retrieves a paginated list of users.
	// It supports cursor-based pagination and filtering.
	// It returns domain.ErrInvalidPaginationCursor if an invalid cursor is provided.
	// It returns domain.ErrInternal if it fails to fetch from the repository.
	ListUsers(ctx context.Context, req ListUsersRequest) (users []*domain.User, nextCursor string, err error)
}

type userUseCaseQueries struct {
	l    logger.Interface
	repo domain.UserRepoQueries
}

func NewUserUseCaseQueries(logger logger.Interface, repo domain.UserRepoQueries) *userUseCaseQueries {
	return &userUseCaseQueries{logger, repo}
}

// GetUser retrieves a single User based on his id.
// It implements the ListUsers method of UserQueries interface
func (uc userUseCaseQueries) GetUser(ctx context.Context, userID string) (du *domain.User, _ error) {
	if _, err := uuid.Parse(userID); err != nil {
		return du, domain.ErrInvalidUserID
	}

	du, err := uc.repo.GetUser(ctx, userID)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			uc.l.Warn("App-user-queries error getting user %s: %v", userID, err)
			return du, domain.ErrInternal
		}
		return du, err
	}
	return du, nil
}

// ListUsers retrieves a paginated list of users.
// It implements the ListUsers method of UserQueries interface
func (uc userUseCaseQueries) ListUsers(ctx context.Context, req ListUsersRequest) (users []*domain.User, nextCur string, err error) {
	var updatedAtCur *time.Time
	var userIDCur string
	if len(req.Cursor) > 0 {
		var uCur time.Time
		uCur, userIDCur, err = decodeCursor(req.Cursor)
		if err != nil {
			uc.l.Debug("App-user-queries error decoding cursor: %v", err)
			return []*domain.User{}, "", domain.ErrInvalidPaginationCursor
		}
		updatedAtCur = &uCur
	}
	users, err = uc.repo.ListUsers(ctx, userIDCur, updatedAtCur, req.Limit, req.Filters)
	if err != nil {
		uc.l.Debug("App-user-queries error list users: %v", err)
		return []*domain.User{}, "", domain.ErrInternal
	}

	if len(users) == int(req.Limit) {
		lastUser := users[len(users)-1]
		nextCur = encodeCursor(lastUser.UpdatedAt, lastUser.ID.String())
	}

	return
}
