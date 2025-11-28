package repo

import (
	"context"
	"database/sql"
	"fmt"
	"friend-help/internal/errs"
	"friend-help/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

type mysqlAuthRepo struct {
	db *sql.DB
}

func NewmysqlAuthRepo(db *sql.DB) AuthRepo {
	return &mysqlAuthRepo{db: db}
}

type AuthRepo interface {
	CheckUserExists(ctx context.Context, login, email string) (bool, error)
	CreateUser(ctx context.Context, user model.AuthUser) (int, error)
	GetUserByLoginOrEmail(ctx context.Context, identifier string) (*model.AuthUser, error)
}

func (r *mysqlAuthRepo) CheckUserExists(ctx context.Context, login, email string) (bool, error) {
	var count int
	var query string
	var args []interface{}
	if email == "" {
		query = "SELECT COUNT(*) FROM users WHERE login = ?"
		args = append(args, login)
	} else {
		query = "SELECT COUNT(*) FROM users WHERE login = ? OR email = ?"
		args = append(args, login, email)
	}
	row := r.db.QueryRowContext(ctx, query, args...)
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to execute CheckUserExists query: %w", err)
	}
	return count > 0, nil
}

func (r *mysqlAuthRepo) CreateUser(ctx context.Context, user model.AuthUser) (int, error) {
	query := `
		INSERT INTO users (username, email, login, password_hash, is_activated)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Login,
		user.PasswordHash,
		user.IsActivated,
	)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errs.ErrDBInsertFailed, err)
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	return int(lastID), nil
}

func (r *mysqlAuthRepo) GetUserByLoginOrEmail(ctx context.Context, identifier string) (*model.AuthUser, error) {
	user := &model.AuthUser{}
	query := `
		SELECT id, login, email, password_hash, is_activated 
		FROM users
		WHERE login = ? OR email = ?
	`
	err := r.db.QueryRowContext(ctx, query, identifier, identifier).Scan(
		&user.ID,
		&user.Login,
		&user.Email,
		&user.PasswordHash,
		&user.IsActivated,
	)
	if err == sql.ErrNoRows {
		return nil, errs.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return user, nil
}
