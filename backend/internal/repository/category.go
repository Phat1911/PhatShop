package repository

import (
	"context"
	"errors"
	"phatshop-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo struct {
	pool *pgxpool.Pool
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

func (r *CategoryRepo) ListAll(ctx context.Context) ([]models.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, slug, product_type, created_at FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ProductType, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	if cats == nil {
		cats = []models.Category{}
	}
	return cats, rows.Err()
}

func (r *CategoryRepo) GetByID(ctx context.Context, id string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, slug, product_type, created_at FROM categories WHERE id=$1`, id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.ProductType, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) Create(ctx context.Context, name, slug, productType string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (name, slug, product_type) VALUES ($1,$2,$3)
		 RETURNING id, name, slug, product_type, created_at`,
		name, slug, productType,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.ProductType, &c.CreatedAt)
	return &c, err
}

func (r *CategoryRepo) Update(ctx context.Context, id, name, slug, productType string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx,
		`UPDATE categories SET name=$2, slug=$3, product_type=$4
		 WHERE id=$1 RETURNING id, name, slug, product_type, created_at`,
		id, name, slug, productType,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.ProductType, &c.CreatedAt)
	return &c, err
}

func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	return err
}
