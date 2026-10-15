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

var _ ports.IDropboxLinkRepository = (*SQLiteDropboxLinkRepository)(nil)

type ITokenCipher interface {
	Encrypt(plain string) (string, error)
	Decrypt(encrypted string) (string, error)
}

// SQLiteDropboxLinkRepository stores the Dropbox tokens encrypted with cipher.
type SQLiteDropboxLinkRepository struct {
	db     *sql.DB
	cipher ITokenCipher
}

func NewSQLiteDropboxLinkRepository(db *sql.DB, cipher ITokenCipher) *SQLiteDropboxLinkRepository {
	return &SQLiteDropboxLinkRepository{db: db, cipher: cipher}
}

func (r *SQLiteDropboxLinkRepository) Get(ctx context.Context, userId int) (*models.DropboxLink, error) {
	var (
		link            = models.DropboxLink{UserId: userId}
		refreshEnc      string
		accessEnc       sql.NullString
		accessExpiresAt sql.NullInt64
	)
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, `
		SELECT accountId, email, refreshTokenEnc, accessTokenEnc, accessExpiresAt, folder
		FROM dropbox_links WHERE userId = ?`, userId,
	).Scan(&link.AccountId, &link.Email, &refreshEnc, &accessEnc, &accessExpiresAt, &link.Folder)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrDropboxNotLinked
	}
	if err != nil {
		return nil, fmt.Errorf("select dropbox link: %w", err)
	}

	if link.RefreshToken, err = r.cipher.Decrypt(refreshEnc); err != nil {
		return nil, fmt.Errorf("decrypt dropbox refresh token: %w", err)
	}
	if accessEnc.Valid {
		if link.Access.Token, err = r.cipher.Decrypt(accessEnc.String); err != nil {
			return nil, fmt.Errorf("decrypt dropbox access token: %w", err)
		}
		link.Access.ExpiresAt = fromNullUnix(accessExpiresAt)
	}
	return &link, nil
}

func (r *SQLiteDropboxLinkRepository) Save(ctx context.Context, link *models.DropboxLink) error {
	refreshEnc, err := r.cipher.Encrypt(link.RefreshToken)
	if err != nil {
		return fmt.Errorf("encrypt dropbox refresh token: %w", err)
	}
	accessEnc, err := r.encryptAccess(link.Access)
	if err != nil {
		return err
	}

	_, err = datasources.GetDB(ctx, r.db).ExecContext(ctx, `
		INSERT INTO dropbox_links (userId, accountId, email, refreshTokenEnc, accessTokenEnc, accessExpiresAt, folder)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (userId) DO UPDATE SET
			accountId = excluded.accountId,
			email = excluded.email,
			refreshTokenEnc = excluded.refreshTokenEnc,
			accessTokenEnc = excluded.accessTokenEnc,
			accessExpiresAt = excluded.accessExpiresAt`,
		link.UserId, link.AccountId, link.Email, refreshEnc, accessEnc, nullUnix(link.Access.ExpiresAt), link.Folder,
	)
	if err != nil {
		return fmt.Errorf("save dropbox link: %w", err)
	}
	return nil
}

func (r *SQLiteDropboxLinkRepository) UpdateAccess(ctx context.Context, userId int, access models.AccessToken) error {
	accessEnc, err := r.encryptAccess(access)
	if err != nil {
		return err
	}
	result, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`UPDATE dropbox_links SET accessTokenEnc = ?, accessExpiresAt = ? WHERE userId = ?`,
		accessEnc, nullUnix(access.ExpiresAt), userId,
	)
	return checkAffected(result, err, models.ErrDropboxNotLinked)
}

func (r *SQLiteDropboxLinkRepository) UpdateFolder(ctx context.Context, userId int, folder string) error {
	result, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`UPDATE dropbox_links SET folder = ? WHERE userId = ?`, folder, userId,
	)
	return checkAffected(result, err, models.ErrDropboxNotLinked)
}

func (r *SQLiteDropboxLinkRepository) Delete(ctx context.Context, userId int) error {
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, `DELETE FROM dropbox_links WHERE userId = ?`, userId)
	if err != nil {
		return fmt.Errorf("delete dropbox link: %w", err)
	}
	return nil
}

func (r *SQLiteDropboxLinkRepository) encryptAccess(access models.AccessToken) (sql.NullString, error) {
	if access.Token == "" {
		return sql.NullString{}, nil
	}
	enc, err := r.cipher.Encrypt(access.Token)
	if err != nil {
		return sql.NullString{}, fmt.Errorf("encrypt dropbox access token: %w", err)
	}
	return sql.NullString{String: enc, Valid: true}, nil
}
