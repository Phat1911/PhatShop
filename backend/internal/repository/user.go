package repository

import (
	"context"
	"errors"
	"phatshop-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, username, email, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)
		 RETURNING id, username, email, password_hash, display_name, avatar_url, role, is_active, created_at, updated_at`,
		username, email, passwordHash,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, display_name, avatar_url, role, is_active, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, password_hash, display_name, avatar_url, role, is_active, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, id, displayName, avatarURL string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET display_name=$2, avatar_url=$3, updated_at=NOW()
		 WHERE id=$1
		 RETURNING id, username, email, password_hash, display_name, avatar_url, role, is_active, created_at, updated_at`,
		id, displayName, avatarURL,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (r *UserRepo) ListAll(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	var total int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)

	rows, err := r.pool.Query(ctx,
		`SELECT id, username, email, display_name, avatar_url, role, is_active, created_at, updated_at
		 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.AvatarURL,
			&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}
	return users, total, rows.Err()
}

func (r *UserRepo) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET role=$2, updated_at=NOW() WHERE id=$1`, id, role)
	return err
}
