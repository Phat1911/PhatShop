package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"phatshop-backend/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DownloadRepo struct {
	pool *pgxpool.Pool
}

func NewDownloadRepo(pool *pgxpool.Pool) *DownloadRepo {
	return &DownloadRepo{pool: pool}
}

func (r *DownloadRepo) HasPurchased(ctx context.Context, userID, productID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM purchases WHERE user_id=$1 AND product_id=$2)`,
		userID, productID,
	).Scan(&exists)
	return exists, err
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (r *DownloadRepo) CreateToken(ctx context.Context, userID, productID string) (*models.DownloadToken, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(15 * time.Minute)

	var dt models.DownloadToken
	err = r.pool.QueryRow(ctx, `
		INSERT INTO download_tokens (user_id, product_id, token, expires_at)
		VALUES ($1,$2,$3,$4)
		RETURNING id, user_id, product_id, token, expires_at, used_count, max_uses, created_at`,
		userID, productID, token, expiresAt,
	).Scan(&dt.ID, &dt.UserID, &dt.ProductID, &dt.Token,
		&dt.ExpiresAt, &dt.UsedCount, &dt.MaxUses, &dt.CreatedAt)
	return &dt, err
}

type TokenWithFile struct {
	TokenID   string
	ProductID string
	FilePath  string
	FileName  string
	UsedCount int
	MaxUses   int
	ExpiresAt time.Time
}

func (r *DownloadRepo) ValidateAndGetFile(ctx context.Context, token string) (*TokenWithFile, error) {
	var t TokenWithFile
	err := r.pool.QueryRow(ctx, `
		SELECT dt.id, dt.product_id, p.file_path, p.file_name,
		       dt.used_count, dt.max_uses, dt.expires_at
		FROM download_tokens dt
		JOIN products p ON p.id = dt.product_id
		WHERE dt.token=$1`, token,
	).Scan(&t.TokenID, &t.ProductID, &t.FilePath, &t.FileName,
		&t.UsedCount, &t.MaxUses, &t.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *DownloadRepo) IncrementUsage(ctx context.Context, tokenID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE download_tokens SET used_count = used_count + 1 WHERE id=$1`, tokenID)
	return err
}
