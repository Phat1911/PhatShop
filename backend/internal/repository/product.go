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
	offset := (f.Page - 1) * f.Limit

	where := []string{}
	args := []interface{}{}
	idx := 1

	if !f.AdminView {
		where = append(where, "p.is_published = TRUE")
	}
	if f.ProductType != "" {
		where = append(where, fmt.Sprintf("p.product_type = $%d", idx))
		args = append(args, f.ProductType)
		idx++
	}
	if f.CategoryID != "" {
		where = append(where, fmt.Sprintf("p.category_id = $%d", idx))
		args = append(args, f.CategoryID)
		idx++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(p.title ILIKE $%d OR p.description ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Search+"%")
		idx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	orderBy := "p.created_at DESC"
	switch f.Sort {
	case "oldest":
		orderBy = "p.created_at ASC"
	case "price_asc":
		orderBy = "p.price ASC"
	case "price_desc":
		orderBy = "p.price DESC"
	case "popular":
		orderBy = "p.purchase_count DESC"
	}

	var total int
	r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM products p %s`, whereClause), args...).Scan(&total)

	args = append(args, f.Limit, offset)
	query := fmt.Sprintf(`
		SELECT p.id, p.seller_id, p.category_id, p.title, p.slug, p.description,
		       p.product_type, p.price, p.thumbnail_url, p.preview_urls,
		       p.file_path, p.file_name, p.file_size, p.tags,
		       p.is_published, p.view_count, p.purchase_count, p.created_at, p.updated_at,
		       COALESCE(u.username, '') as seller_name,
		       COALESCE(c.name, '') as category_name
		FROM products p
		LEFT JOIN users u ON u.id = p.seller_id
		LEFT JOIN categories c ON c.id = p.category_id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.SellerID, &p.CategoryID, &p.Title, &p.Slug, &p.Description,
			&p.ProductType, &p.Price, &p.ThumbnailURL, &p.PreviewURLs,
			&p.FilePath, &p.FileName, &p.FileSize, &p.Tags,
			&p.IsPublished, &p.ViewCount, &p.PurchaseCount, &p.CreatedAt, &p.UpdatedAt,
			&p.SellerName, &p.CategoryName,
		); err != nil {
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
		SELECT p.id, p.seller_id, p.category_id, p.title, p.slug, p.description,
		       p.product_type, p.price, p.thumbnail_url, p.preview_urls,
		       p.file_path, p.file_name, p.file_size, p.tags,
		       p.is_published, p.view_count, p.purchase_count, p.created_at, p.updated_at,
		       COALESCE(u.username, '') as seller_name,
		       COALESCE(c.name, '') as category_name
		FROM products p
		LEFT JOIN users u ON u.id = p.seller_id
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1`, id,
	).Scan(
		&p.ID, &p.SellerID, &p.CategoryID, &p.Title, &p.Slug, &p.Description,
		&p.ProductType, &p.Price, &p.ThumbnailURL, &p.PreviewURLs,
		&p.FilePath, &p.FileName, &p.FileSize, &p.Tags,
		&p.IsPublished, &p.ViewCount, &p.PurchaseCount, &p.CreatedAt, &p.UpdatedAt,
		&p.SellerName, &p.CategoryName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *models.Product) (*models.Product, error) {
	var created models.Product
	err := r.pool.QueryRow(ctx, `
		INSERT INTO products (seller_id, category_id, title, slug, description, product_type, price,
		                      thumbnail_url, preview_urls, file_path, file_name, file_size, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, seller_id, category_id, title, slug, description, product_type, price,
		          thumbnail_url, preview_urls, file_path, file_name, file_size, tags,
		          is_published, view_count, purchase_count, created_at, updated_at`,
		p.SellerID, p.CategoryID, p.Title, p.Slug, p.Description, p.ProductType, p.Price,
		p.ThumbnailURL, p.PreviewURLs, p.FilePath, p.FileName, p.FileSize, p.Tags,
	).Scan(
		&created.ID, &created.SellerID, &created.CategoryID, &created.Title, &created.Slug, &created.Description,
		&created.ProductType, &created.Price, &created.ThumbnailURL, &created.PreviewURLs,
		&created.FilePath, &created.FileName, &created.FileSize, &created.Tags,
		&created.IsPublished, &created.ViewCount, &created.PurchaseCount, &created.CreatedAt, &created.UpdatedAt,
	)
	created.FilePath = ""
	return &created, err
}

func (r *ProductRepo) Update(ctx context.Context, id string, p *models.Product) (*models.Product, error) {
	var updated models.Product
	err := r.pool.QueryRow(ctx, `
		UPDATE products SET category_id=$2, title=$3, slug=$4, description=$5,
		                    product_type=$6, price=$7, tags=$8, updated_at=NOW()
		WHERE id=$1
		RETURNING id, seller_id, category_id, title, slug, description, product_type, price,
		          thumbnail_url, preview_urls, file_path, file_name, file_size, tags,
		          is_published, view_count, purchase_count, created_at, updated_at`,
		id, p.CategoryID, p.Title, p.Slug, p.Description, p.ProductType, p.Price, p.Tags,
	).Scan(
		&updated.ID, &updated.SellerID, &updated.CategoryID, &updated.Title, &updated.Slug, &updated.Description,
		&updated.ProductType, &updated.Price, &updated.ThumbnailURL, &updated.PreviewURLs,
		&updated.FilePath, &updated.FileName, &updated.FileSize, &updated.Tags,
		&updated.IsPublished, &updated.ViewCount, &updated.PurchaseCount, &updated.CreatedAt, &updated.UpdatedAt,
	)
	updated.FilePath = ""
	return &updated, err
}

func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

func (r *ProductRepo) SetPublished(ctx context.Context, id string, published bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE products SET is_published=$2, updated_at=NOW() WHERE id=$1`, id, published)
	return err
}

func (r *ProductRepo) IncrementView(ctx context.Context, id string) {
	r.pool.Exec(ctx, `UPDATE products SET view_count = view_count + 1 WHERE id=$1`, id)
}

func (r *ProductRepo) GetFilePath(ctx context.Context, id string) (string, string, error) {
	var filePath, fileName string
	err := r.pool.QueryRow(ctx, `SELECT file_path, file_name FROM products WHERE id=$1`, id).Scan(&filePath, &fileName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return filePath, fileName, nil
}
