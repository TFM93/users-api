package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	// UserRepoCommands is an interface for persisting users
	UserRepoCommands interface {
		// SaveUser creates a new user in the database.
		// Returns the created User's ID and an error in case of failure.
		// If user already exists or a conflict is found, it returns domain.ErrUserAlreadyExists.
		// If an internal error occurs, it logs the error and returns domain.ErrInternal.
		SaveUser(ctx context.Context, user *User) (string, error)

		// DeleteUser deletes user by their ID from the database.
		// If user does not exist, it returns domain.ErrUserNotFound.
		// If an internal error occurs, it logs the error and returns domain.ErrInternal.
		DeleteUser(ctx context.Context, userID string) error

		// UpdateUser updates user in database.
		// If user does not exist, it returns domain.ErrUserNotFound.
		// If an internal error occurs, it logs the error and returns domain.ErrInternal.
		UpdateUser(ctx context.Context, user *User) error
	}

	// UserRepoQueries is an interface for query persisted users
	UserRepoQueries interface {
		// GetUser fetches a single user from the database based on the userID.
		// Returns the user object and an error if the operation fails.
		// If the user does not exist, it returns domain.ErrUserNotFound.
		// If there's an error processing the data, it returns domain.ErrFailedToProcessData.
		GetUser(ctx context.Context, userID string) (*User, error)

		// ListUsers fetches a list of users based on the provided filters and pagination options.
		// Parameters:
		//   cursorUserID: ID of the user to start listing from
		//   cursorUpdatedAt: Timestamp to filter users updated after this time
		//   limit: Maximum number of users to return
		//   filters: Criteria to filter users by
		// Returns a slice of user objects and an error if the operation fails.
		// If the query fails to execute, it returns return domain.ErrInternal.
		// If theres an error processing the data, it returns domain.ErrFailedToProcessData.
		ListUsers(ctx context.Context, cursorUserID string, cursorUpdatedAt *time.Time, limit int32, filters UserSearchFilters) ([]*User, error)
	}

	// User represents a User in the domain model
	User struct {
		ID             uuid.UUID
		FirstName      string
		LastName       string
		NickName       string
		Email          string
		Password       string
		CountryISOCode string
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}

	// UserSearchFilters represents User's searchable fields
	UserSearchFilters struct {
		FirstName      *string
		LastName       *string
		NickName       *string
		Email          *string
		CountryISOCode *string
	}
)
