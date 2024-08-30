package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"
	"users/internal/domain"
	log "users/pkg/logger"
	"users/pkg/postgresql"
)

type userQueriesRepo struct {
	pg postgresql.Interface
	l  log.Interface
}

// NewUserQueriesRepo creates a new instance of userQueriesRepo that satisfies the domain.UserRepoQueries interface
func NewUserQueriesRepo(pg postgresql.Interface, logger log.Interface) domain.UserRepoQueries {
	ur := &userQueriesRepo{pg, logger}
	return ur
}

func (r userQueriesRepo) db(ctx context.Context) postgresql.DBProvider {
	tx, ok := ctx.Value(domain.TxKey).(postgresql.Tx)
	if ok {
		return tx
	}
	return r.pg.GetPool()
}

// GetUser fetches a single user from the database based on the userID
// If theres an error processing the data, it returns domain.ErrFailedToProcessData
func (r userQueriesRepo) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	query := `SELECT id, first_name, last_name, country_iso_code, nickname, email, created_at, updated_at FROM users WHERE id = $1`
	row := r.db(ctx).QueryRow(ctx, query, userID)
	var user domain.User
	if err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.CountryISOCode, &user.NickName, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == postgresql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		r.l.Error(fmt.Errorf("failed to scan row: %w", err))
		return nil, domain.ErrFailedToProcessData
	}
	return &user, nil
}

// ListUsers fetches the users from the database based on the provided cursor data
// If the query fails to execute, it returns return domain.ErrInternal
// If theres an error processing the data, it returns domain.ErrFailedToProcessData
func (r userQueriesRepo) ListUsers(ctx context.Context, cursorUserID string, cursorUpdatedAt *time.Time, limit int32, filters domain.UserSearchFilters) ([]*domain.User, error) {
	var whereClauses []string
	var args []any

	// pagination
	if cursorUserID != "" && cursorUpdatedAt != nil {
		whereClauses = append(whereClauses, `(updated_at < $1 OR (updated_at = $1 AND id < $2))`)
		args = append(args, *cursorUpdatedAt, cursorUserID)
	}

	// filters
	if filters.FirstName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("first_name ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filters.FirstName+"%")
	}
	if filters.LastName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("last_name ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filters.LastName+"%")
	}
	if filters.NickName != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("nickname ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filters.NickName+"%")
	}
	if filters.Email != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("email ILIKE $%d", len(args)+1))
		args = append(args, "%"+*filters.Email+"%")
	}
	if filters.CountryISOCode != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("country_iso_code ILIKE $%d", len(args)+1))
		args = append(args, *filters.CountryISOCode)
	}

	// compose query statement
	where := ""
	if len(whereClauses) > 0 {
		where = "WHERE " + strings.Join(whereClauses, " AND ")
	}
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, country_iso_code, nickname, email, created_at, updated_at 
		FROM users 
		%s 
		ORDER BY updated_at DESC, id DESC LIMIT $%d`, where, len(args)+1)
	args = append(args, limit)

	rows, err := r.db(ctx).Query(ctx, query, args...)
	if err != nil {
		r.l.Debug(fmt.Errorf("failed to list users: %w", err))
		return nil, domain.ErrInternal
	}
	defer rows.Close()

	var users []*domain.User
	// NOTE: as per version 5 of pgx this can be done relying on generics:
	// https://donchev.is/post/working-with-postgresql-in-go-using-pgx/
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.CountryISOCode, &user.NickName, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			r.l.Error(fmt.Errorf("failed to scan row: %w", err))
			return nil, domain.ErrFailedToProcessData
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.l.Error(fmt.Errorf("row iteration error: %w", err))
		return nil, domain.ErrFailedToProcessData
	}
	return users, nil
}
