package postgresql

import (
	"context"
	"fmt"
	"users/internal/domain"
	log "users/pkg/logger"
	"users/pkg/postgresql"
)

type userCommandsRepo struct {
	pg postgresql.Interface
	l  log.Interface
}

// NewUserCommandsRepo creates a new instance of userCommandsRepo that satisfies the domain.UserRepoCommands interface
func NewUserCommandsRepo(pg postgresql.Interface, logger log.Interface) domain.UserRepoCommands {
	ur := &userCommandsRepo{pg: pg, l: logger}
	return ur
}

func (r userCommandsRepo) db(ctx context.Context) postgresql.DBProvider {
	tx, ok := ctx.Value(domain.TxKey).(postgresql.Tx)
	if ok {
		return tx
	}
	return r.pg.GetPool()
}

// SaveUser creates a new user in the database.
// If user already exists or a conflict is found, it returns domain.ErrUserAlreadyExists
// If an internal error occurs, it logs the error and returns domain.ErrInternal
func (r userCommandsRepo) SaveUser(ctx context.Context, user *domain.User) (id string, err error) {
	query := `INSERT INTO users (first_name, last_name, country_iso_code, nickname, email, pw) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = r.db(ctx).QueryRow(ctx, query,
		user.FirstName,
		user.LastName,
		user.CountryISOCode,
		user.NickName,
		user.Email,
		user.Password).Scan(&id)
	if err != nil {
		if postgresql.IsConflictErr(err) {
			r.l.Debug(fmt.Errorf("user %s already exists: %w", user.Email, err))
			return "", domain.ErrUserAlreadyExists
		}

		r.l.Error(fmt.Errorf("failed to save user: %w", err))
		return id, domain.ErrInternal
	}
	return id, nil
}

// DeleteUser deletes user by their ID from the database.
// If user does not exist, it returns domain.ErrUserNotFound
// If an internal error occurs, it logs the error and returns domain.ErrInternal
func (r userCommandsRepo) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id=$1`
	commandTag, err := r.db(ctx).Exec(ctx, query, userID)
	if err != nil {
		r.l.Error(fmt.Errorf("failed to delete user: %w", err))
		return domain.ErrInternal
	}
	if commandTag.RowsAffected() == 0 {
		r.l.Debug("user with ID %s does not exist", userID)
		return domain.ErrUserNotFound
	}
	return err
}

// UpdateUser updates user in database.
// If user does not exist, it returns domain.ErrUserNotFound
// If an internal error occurs, it logs the error and returns domain.ErrInternal
func (r userCommandsRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET first_name=$2, last_name=$3, country_iso_code=$4, nickname=$5, email=$6 WHERE id=$1;`
	commandTag, err := r.db(ctx).Exec(ctx, query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.CountryISOCode,
		user.NickName,
		user.Email)
	if err != nil {
		r.l.Error(fmt.Errorf("failed to update user: %w", err))
		return domain.ErrInternal
	}
	if commandTag.RowsAffected() == 0 {
		r.l.Debug("user with ID %s does not exist", user.ID)
		return domain.ErrUserNotFound
	}
	return err
}
