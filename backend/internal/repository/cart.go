package repository

import (
	"context"
	"phatshop-backend/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CartRepo struct {
	pool *pgxpool.Pool
}

func NewCartRepo(pool *pgxpool.Pool) *CartRepo {
	return &CartRepo{pool: pool}
}

func (r *CartRepo) GetItems(ctx context.Context, userID string) ([]models.CartItem, int64, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ci.id, ci.user_id, ci.product_id, ci.created_at,
		       p.id, p.seller_id, p.category_id, p.title, p.slug, p.description, p.product_type,
		       p.price, p.thumbnail_url, p.preview_urls, p.file_path, p.file_name, p.file_size,
		       p.tags, p.is_published, p.view_count, p.purchase_count, p.trailer_url, p.created_at, p.updated_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.user_id = $1
		ORDER BY ci.created_at DESC`, userID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.CartItem
	var total int64
	for rows.Next() {
		var ci models.CartItem
		var p models.Product
		err := rows.Scan(
			&ci.ID, &ci.UserID, &ci.ProductID, &ci.CreatedAt,
			&p.ID, &p.SellerID, &p.CategoryID, &p.Title, &p.Slug, &p.Description, &p.ProductType,
			&p.Price, &p.ThumbnailURL, &p.PreviewURLs, &p.FilePath, &p.FileName, &p.FileSize,
			&p.Tags, &p.IsPublished, &p.ViewCount, &p.PurchaseCount, &p.TrailerURL, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		p.FilePath = ""
		ci.Product = &p
		total += p.Price
		items = append(items, ci)
	}
	if items == nil {
		items = []models.CartItem{}
	}
	return items, total, rows.Err()
}

func (r *CartRepo) AddItem(ctx context.Context, userID, productID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO cart_items (user_id, product_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		userID, productID)
	return err
}

func (r *CartRepo) RemoveItem(ctx context.Context, userID, productID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM cart_items WHERE user_id=$1 AND product_id=$2`, userID, productID)
	return err
}

func (r *CartRepo) Clear(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE user_id=$1`, userID)
	return err
}
