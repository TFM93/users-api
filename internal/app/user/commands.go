package user

import (
	"context"
	"encoding/json"
	"errors"
	"unicode"
	"users/internal/domain"
	"users/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserCommands interface {
	// CreateUser creates a new User and returns the created user id.
	// It returns domain.ErrInvalidPW if an invalid password is provided.
	// It returns domain.ErrUserAlreadyExists if the user conflicts in the unique fields (email or nickname).
	// It returns domain.ErrInternal if it fails to create.
	CreateUser(ctx context.Context, req AddUserRequest) (userID string, err error)

	// UpdateUser updates a single User based on his id.
	// This is not a partial update, all the user fields should be provided.
	// It returns domain.ErrInvalidUserID if an invalid user id is provided.
	// It returns domain.ErrUserNotFound if the user does not exist.
	// It returns domain.ErrInternal if it fails to update.
	UpdateUser(ctx context.Context, req UpdateUserRequest) error

	// DeleteUser deletes a single User based on his id.
	// It returns domain.ErrInvalidUserID if an invalid user id is provided.
	// It returns domain.ErrUserNotFound if the user does not exist.
	// It returns domain.ErrInternal if it fails to delete.
	DeleteUser(ctx context.Context, userID string) error
}

type AddUserRequest struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	NickName       string `json:"nickname"`
	Email          string `json:"email"`
	Password       string `json:"-"`
	CountryISOCode string `json:"country"`
}

type UpdateUserRequest struct {
	ID             string `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	NickName       string `json:"nickname"`
	Email          string `json:"email"`
	CountryISOCode string `json:"country"`
}

type userUseCaseCommands struct {
	l           logger.Interface
	repo        domain.UserRepoCommands
	outboxRepo  domain.OutboxRepoCommands
	transaction domain.Transaction
}

func NewUserUseCaseCommands(logger logger.Interface, repo domain.UserRepoCommands, transaction domain.Transaction, outboxRepo domain.OutboxRepoCommands) *userUseCaseCommands {
	return &userUseCaseCommands{logger, repo, outboxRepo, transaction}
}

// CreateUser creates a new User and returns the created user id.
// It implements the CreateUser method of UserCommands interface
func (uc userUseCaseCommands) CreateUser(ctx context.Context, req AddUserRequest) (string, error) {
	if err := validatePassword(req.Password); err != nil {
		return "", err
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		uc.l.Debug("app-user-commands-create - password hashing error: %s", err)
		return "", domain.ErrInvalidPW
	}

	u := domain.User{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NickName:       req.NickName,
		Email:          req.Email,
		CountryISOCode: req.CountryISOCode,
		Password:       hashedPassword,
	}

	var userID string

	if err := uc.transaction.BeginTx(ctx, func(txCtx context.Context) error {
		userID, err = uc.repo.SaveUser(txCtx, &u)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(req)
		if err != nil {
			return err
		}
		event := &domain.Event{
			Type:    "CreateUser",
			Payload: payload,
		}
		if _, err := uc.outboxRepo.AddEvent(txCtx, event); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if !errors.Is(err, domain.ErrUserAlreadyExists) {
			uc.l.Warn("app-user-commands-create error: %v", err)
			return "", domain.ErrInternal
		}
		return "", err
	}
	return userID, nil
}

// DeleteUser deletes a single User based on his id.
// It implements the DeleteUser method of UserCommands interface
func (uc userUseCaseCommands) DeleteUser(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return domain.ErrInvalidUserID
	}

	if err := uc.transaction.BeginTx(ctx, func(txCtx context.Context) error {
		err := uc.repo.DeleteUser(ctx, userID)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(struct{ ID string }{
			ID: userID,
		})
		if err != nil {
			return err
		}
		event := &domain.Event{
			Type:    "DeleteUser",
			Payload: payload,
		}
		if _, err := uc.outboxRepo.AddEvent(txCtx, event); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			uc.l.Warn("app-user-commands-delete error: %v", err)
			return domain.ErrInternal
		}
		return err
	}
	return nil
}

// UpdateUser updates a single User based on his id.
// It implements the UpdateUser method of UserCommands interface
func (uc userUseCaseCommands) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	userID, err := uuid.Parse(req.ID)
	if err != nil {
		return domain.ErrInvalidUserID
	}

	u := domain.User{
		ID:             userID,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		NickName:       req.NickName,
		Email:          req.Email,
		CountryISOCode: req.CountryISOCode,
	}

	if err := uc.transaction.BeginTx(ctx, func(txCtx context.Context) error {
		err = uc.repo.UpdateUser(ctx, &u)
		if err != nil {
			return err
		}

		payload, err := json.Marshal(req)
		if err != nil {
			return err
		}

		event := &domain.Event{
			Type:    "UpdateUser",
			Payload: payload,
		}
		if _, err := uc.outboxRepo.AddEvent(txCtx, event); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			uc.l.Warn("app-user-commands-update error updating user %s: %v", req.ID, err)
			return domain.ErrInternal
		}
		return err
	}
	return nil
}

// validatePassword does a very basic password validation
func validatePassword(password string) error {
	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit || len(password) < 6 {
		return domain.ErrInvalidPW
	}
	return nil
}

// hashPassword uses bcrypt to generate a password hash using DefaultCost
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
