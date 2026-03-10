package repository

import (
	"context"
	"errors"
	"fmt"
	"phatshop-backend/internal/models"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

type ProductFilter struct {
	ProductType string
	CategoryID  string
	Search      string
	Sort        string
	Page        int
	Limit       int
	AdminView   bool
}

func (r *ProductRepo) List(ctx context.Context, f ProductFilter) ([]models.Product, int, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "p.is_published = true")

	if f.ProductType != "" {
		conditions = append(conditions, fmt.Sprintf("p.product_type = $%d", argIdx))
		args = append(args, f.ProductType)
		argIdx++
	}
	if f.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argIdx))
		args = append(args, f.CategoryID)
		argIdx++
	}
	if f.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(p.title ILIKE $%d OR p.description ILIKE $%d OR $%d = ANY(p.tags))",
			argIdx, argIdx+1, argIdx+2))
		like := "%" + f.Search + "%"
		args = append(args, like, like, f.Search)
		argIdx += 3
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderBy := "p.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderBy = "p.price ASC"
	case "price_desc":
		orderBy = "p.price DESC"
	case "popular":
		orderBy = "p.purchase_count DESC"
	case "views":
		orderBy = "p.view_count DESC"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM products p
		LEFT JOIN users u ON u.id = p.seller_id
		%s`, where)

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)

	query := fmt.Sprintf(`
		SELECT p.id, p.seller_id, p.category_id, p.title, p.slug, p.description, p.product_type,
		       p.price, p.thumbnail_url, p.preview_urls, p.file_path, p.file_name, p.file_size,
		       p.tags, p.is_published, p.view_count, p.purchase_count, p.created_at, p.updated_at,
		       COALESCE(u.display_name, u.username, '') as seller_name,
		       COALESCE(c.name, '') as category_name,
		       COALESCE(p.trailer_url, '') as trailer_url
		FROM products p
		LEFT JOIN users u ON u.id = p.seller_id
		LEFT JOIN categories c ON c.id = p.category_id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, where, orderBy, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		err := rows.Scan(
			&p.ID, &p.SellerID, &p.CategoryID, &p.Title, &p.Slug, &p.Description, &p.ProductType,
			&p.Price, &p.ThumbnailURL, &p.PreviewURLs, &p.FilePath, &p.FileName, &p.FileSize,
			&p.Tags, &p.IsPublished, &p.ViewCount, &p.PurchaseCount, &p.CreatedAt, &p.UpdatedAt,
			&p.SellerName, &p.CategoryName, &p.TrailerURL,
		)
		if err != nil {
			return nil, 0, err
		}
		p.FilePath = ""
		products = append(products, p)
	}
	if products == nil {
		products = []models.Product{}
	}
	return products, total, rows.Err()
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*models.Product, error) {
	var p models.Product
	err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.seller_id, p.category_id, p.title, p.slug, p.description, p.product_type,
		       p.price, p.thumbnail_url, p.preview_urls, p.file_path, p.file_name, p.file_size,
		       p.tags, p.is_published, p.view_count, p.purchase_count, p.created_at, p.updated_at,
		       COALESCE(u.display_name, u.username, '') as seller_name,
		       COALESCE(c.name, '') as category_name,
		       COALESCE(p.trailer_url, '') as trailer_url
		FROM products p
		LEFT JOIN users u ON u.id = p.seller_id
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1`, id).Scan(
		&p.ID, &p.SellerID, &p.CategoryID, &p.Title, &p.Slug, &p.Description, &p.ProductType,
		&p.Price, &p.ThumbnailURL, &p.PreviewURLs, &p.FilePath, &p.FileName, &p.FileSize,
		&p.Tags, &p.IsPublished, &p.ViewCount, &p.PurchaseCount, &p.CreatedAt, &p.UpdatedAt,
		&p.SellerName, &p.CategoryName, &p.TrailerURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// IncrementView uses context.Background() so it survives after the request context is cancelled.
func (r *ProductRepo) IncrementView(id string) {
	r.pool.Exec(context.Background(), `UPDATE products SET view_count = view_count + 1 WHERE id=$1`, id)
}

// IncrementPurchaseCount increments purchase_count for a list of product IDs.
func (r *ProductRepo) IncrementPurchaseCount(ctx context.Context, productIDs []string) {
	for _, pid := range productIDs {
		r.pool.Exec(context.Background(), `UPDATE products SET purchase_count = purchase_count + 1 WHERE id=$1`, pid)
	}
}

func (r *ProductRepo) GetFilePath(ctx context.Context, id string) (string, string, error) {
	var filePath, fileName string
	err := r.pool.QueryRow(ctx,
		`SELECT file_path, file_name FROM products WHERE id=$1`, id).Scan(&filePath, &fileName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return filePath, fileName, nil
}

// Create inserts a new product record.
func (r *ProductRepo) Create(ctx context.Context, p *models.Product) (*models.Product, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO products
			(id, seller_id, category_id, title, slug, description, product_type,
			 price, thumbnail_url, preview_urls, file_path, file_name, file_size,
			 tags, trailer_url, is_published, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,false,NOW(),NOW())`,
		p.ID, p.SellerID, p.CategoryID, p.Title, p.Slug, p.Description, p.ProductType,
		p.Price, p.ThumbnailURL, p.PreviewURLs, p.FilePath, p.FileName, p.FileSize, p.Tags, p.TrailerURL,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, p.ID)
}

// Delete removes a product and its associated file from storage.
func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

// SetPublished toggles the is_published flag on a product.
func (r *ProductRepo) SetPublished(ctx context.Context, id string, published bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE products SET is_published=$2, updated_at=NOW() WHERE id=$1`, id, published)
	return err
}
