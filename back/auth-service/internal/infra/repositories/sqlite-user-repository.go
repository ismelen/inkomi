package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/datasources"
)

var _ ports.IUserRepository = (*SQLiteUserRepository)(nil)

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

const selectUser = `SELECT id, email, emailVerified, passwordHash, role, fcmToken FROM users`

func (r *SQLiteUserRepository) Create(ctx context.Context, user *models.User) error {
	result, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`INSERT INTO users (email, emailVerified, passwordHash, role, fcmToken) VALUES (?, ?, ?, ?, ?)`,
		nullString(user.Email), user.EmailVerified, nullString(user.PasswordHash), user.Role, nullString(user.FCMToken),
	)
	if isUniqueViolation(err) {
		return models.ErrEmailTaken
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get user id: %w", err)
	}
	user.Id = int(id)
	return nil
}

func (r *SQLiteUserRepository) GetById(ctx context.Context, id int) (*models.User, error) {
	return r.get(ctx, selectUser+` WHERE id = ?`, id)
}

func (r *SQLiteUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.get(ctx, selectUser+` WHERE email = ?`, email)
}

func (r *SQLiteUserRepository) Update(ctx context.Context, user *models.User) error {
	result, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`UPDATE users SET email = ?, emailVerified = ?, passwordHash = ? WHERE id = ?`,
		nullString(user.Email), user.EmailVerified, nullString(user.PasswordHash), user.Id,
	)
	if isUniqueViolation(err) {
		return models.ErrEmailTaken
	}
	return checkAffected(result, err, models.ErrUserNotFound)
}

func (r *SQLiteUserRepository) UpdateFCMToken(ctx context.Context, id int, token string) error {
	result, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`UPDATE users SET fcmToken = ? WHERE id = ?`, nullString(token), id,
	)
	return checkAffected(result, err, models.ErrUserNotFound)
}

func (r *SQLiteUserRepository) get(ctx context.Context, query string, arg any) (*models.User, error) {
	var (
		user                      models.User
		email, password, fcmToken sql.NullString
	)
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, arg).Scan(
		&user.Id, &email, &user.EmailVerified, &password, &user.Role, &fcmToken,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	user.Email, user.PasswordHash, user.FCMToken = email.String, password.String, fcmToken.String
	return &user, nil
}
