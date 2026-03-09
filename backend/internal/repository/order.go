package repository

import (
	"context"
	"errors"
	"fmt"
	"phatshop-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) Create(ctx context.Context, buyerID string, items []models.CartItem) (*models.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var totalAmount int64
	for _, item := range items {
		totalAmount += item.Product.Price
	}

	var order models.Order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (buyer_id, total_amount) VALUES ($1,$2)
		 RETURNING id, buyer_id, total_amount, status, vnpay_txn_ref, vnpay_txn_no,
		           vnpay_bank_code, payment_at, created_at, updated_at`,
		buyerID, totalAmount,
	).Scan(&order.ID, &order.BuyerID, &order.TotalAmount, &order.Status,
		&order.VNPayTxnRef, &order.VNPayTxnNo, &order.VNPayBankCode,
		&order.PaymentAt, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, price) VALUES ($1,$2,$3)`,
			order.ID, item.ProductID, item.Product.Price)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var o models.Order
	err := r.pool.QueryRow(ctx, `
		SELECT o.id, o.buyer_id, o.total_amount, o.status, o.vnpay_txn_ref, o.vnpay_txn_no,
		       o.vnpay_bank_code, o.payment_at, o.created_at, o.updated_at,
		       COALESCE(u.username, '') as buyer_name
		FROM orders o
		LEFT JOIN users u ON u.id = o.buyer_id
		WHERE o.id=$1`, id,
	).Scan(&o.ID, &o.BuyerID, &o.TotalAmount, &o.Status, &o.VNPayTxnRef, &o.VNPayTxnNo,
		&o.VNPayBankCode, &o.PaymentAt, &o.CreatedAt, &o.UpdatedAt, &o.BuyerName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT oi.id, oi.order_id, oi.product_id, oi.price, oi.created_at,
		       p.id, p.seller_id, p.category_id, p.title, p.slug, p.description, p.product_type,
		       p.price, p.thumbnail_url, p.preview_urls, '' as file_path, p.file_name, p.file_size,
		       p.tags, p.is_published, p.view_count, p.purchase_count, p.created_at, p.updated_at
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id=$1`, id)
	if err != nil {
		return &o, nil
	}
	defer rows.Close()
	for rows.Next() {
		var item models.OrderItem
		var p models.Product
		rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Price, &item.CreatedAt,
			&p.ID, &p.SellerID, &p.CategoryID, &p.Title, &p.Slug, &p.Description, &p.ProductType,
			&p.Price, &p.ThumbnailURL, &p.PreviewURLs, &p.FilePath, &p.FileName, &p.FileSize,
			&p.Tags, &p.IsPublished, &p.ViewCount, &p.PurchaseCount, &p.CreatedAt, &p.UpdatedAt,
		)
		item.Product = &p
		o.Items = append(o.Items, item)
	}
	return &o, nil
}

func (r *OrderRepo) GetByTxnRef(ctx context.Context, txnRef string) (*models.Order, error) {
	var o models.Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, buyer_id, total_amount, status, vnpay_txn_ref, vnpay_txn_no,
		       vnpay_bank_code, payment_at, created_at, updated_at
		FROM orders WHERE vnpay_txn_ref=$1`, txnRef,
	).Scan(&o.ID, &o.BuyerID, &o.TotalAmount, &o.Status, &o.VNPayTxnRef, &o.VNPayTxnNo,
		&o.VNPayBankCode, &o.PaymentAt, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, buyerID string) ([]models.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, buyer_id, total_amount, status, vnpay_txn_ref, vnpay_txn_no,
		       vnpay_bank_code, payment_at, created_at, updated_at
		FROM orders WHERE buyer_id=$1 ORDER BY created_at DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []models.Order
	for rows.Next() {
		var o models.Order
		rows.Scan(&o.ID, &o.BuyerID, &o.TotalAmount, &o.Status, &o.VNPayTxnRef, &o.VNPayTxnNo,
			&o.VNPayBankCode, &o.PaymentAt, &o.CreatedAt, &o.UpdatedAt)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []models.Order{}
	}
	return orders, rows.Err()
}

func (r *OrderRepo) ListAll(ctx context.Context, status string, limit, offset int) ([]models.Order, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	idx := 1
	if status != "" {
		where += fmt.Sprintf(" AND o.status = $%d", idx)
		args = append(args, status)
		idx++
	}

	var total int
	r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM orders o "+where, args...).Scan(&total)

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT o.id, o.buyer_id, o.total_amount, o.status, o.vnpay_txn_ref, o.vnpay_txn_no,
		       o.vnpay_bank_code, o.payment_at, o.created_at, o.updated_at,
		       COALESCE(u.username, '') as buyer_name
		FROM orders o LEFT JOIN users u ON u.id=o.buyer_id
		%s ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var orders []models.Order
	for rows.Next() {
		var o models.Order
		rows.Scan(&o.ID, &o.BuyerID, &o.TotalAmount, &o.Status, &o.VNPayTxnRef, &o.VNPayTxnNo,
			&o.VNPayBankCode, &o.PaymentAt, &o.CreatedAt, &o.UpdatedAt, &o.BuyerName)
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []models.Order{}
	}
	return orders, total, rows.Err()
}

func (r *OrderRepo) UpdateTxnRef(ctx context.Context, id, txnRef string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET vnpay_txn_ref=$2, updated_at=NOW() WHERE id=$1`, id, txnRef)
	return err
}

func (r *OrderRepo) MarkPaid(ctx context.Context, txnRef, txnNo, bankCode string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var orderID, buyerID string
	err = tx.QueryRow(ctx,
		`UPDATE orders SET status='paid', vnpay_txn_no=$2, vnpay_bank_code=$3,
		                   payment_at=NOW(), updated_at=NOW()
		 WHERE vnpay_txn_ref=$1 AND status='pending'
		 RETURNING id, buyer_id`, txnRef, txnNo, bankCode,
	).Scan(&orderID, &buyerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO purchases (user_id, product_id, order_id)
		SELECT $1, oi.product_id, oi.order_id
		FROM order_items oi WHERE oi.order_id=$2
		ON CONFLICT DO NOTHING`, buyerID, orderID); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `
		UPDATE products SET purchase_count = purchase_count + 1
		WHERE id IN (SELECT product_id FROM order_items WHERE order_id=$1)`, orderID); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `
		DELETE FROM cart_items WHERE user_id=$1
		AND product_id IN (SELECT product_id FROM order_items WHERE order_id=$2)`,
		buyerID, orderID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	return err
}

// MarkPaidByAdmin is called when the admin manually confirms a bank transfer payment.
// It updates order status to 'paid' and inserts purchases so buyers can download.
func (r *OrderRepo) MarkPaidByAdmin(ctx context.Context, orderID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var buyerID string
	err = tx.QueryRow(ctx,
		`UPDATE orders SET status='paid', payment_at=NOW(), updated_at=NOW()
		 WHERE id=$1 AND status='pending'
		 RETURNING buyer_id`, orderID,
	).Scan(&buyerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Already paid or not found — just update status without side effects
			_, err = tx.Exec(ctx, `UPDATE orders SET status='paid', updated_at=NOW() WHERE id=$1`, orderID)
			if err != nil {
				return err
			}
			return tx.Commit(ctx)
		}
		return err
	}

	// Insert purchase records so buyer can download
	if _, err = tx.Exec(ctx, `
		INSERT INTO purchases (user_id, product_id, order_id)
		SELECT $1, oi.product_id, oi.order_id
		FROM order_items oi WHERE oi.order_id=$2
		ON CONFLICT DO NOTHING`, buyerID, orderID); err != nil {
		return err
	}

	// Increment purchase_count on products
	if _, err = tx.Exec(ctx, `
		UPDATE products SET purchase_count = purchase_count + 1
		WHERE id IN (SELECT product_id FROM order_items WHERE order_id=$1)`, orderID); err != nil {
		return err
	}

	// Remove purchased items from buyer's cart
	if _, err = tx.Exec(ctx, `
		DELETE FROM cart_items WHERE user_id=$1
		AND product_id IN (SELECT product_id FROM order_items WHERE order_id=$2)`,
		buyerID, orderID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrderRepo) GetStats(ctx context.Context) (models.AdminStats, error) {
	var stats models.AdminStats
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE is_published=TRUE`).Scan(&stats.TotalProducts)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&stats.TotalOrders)
	r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_amount),0) FROM orders WHERE status='paid'`).Scan(&stats.TotalRevenue)
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)

	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.buyer_id, o.total_amount, o.status, o.vnpay_txn_ref, o.vnpay_txn_no,
		       o.vnpay_bank_code, o.payment_at, o.created_at, o.updated_at,
		       COALESCE(u.username, '') as buyer_name
		FROM orders o LEFT JOIN users u ON u.id=o.buyer_id
		ORDER BY o.created_at DESC LIMIT 10`)
	if err != nil {
		return stats, nil
	}
	defer rows.Close()
	for rows.Next() {
		var o models.Order
		rows.Scan(&o.ID, &o.BuyerID, &o.TotalAmount, &o.Status, &o.VNPayTxnRef, &o.VNPayTxnNo,
			&o.VNPayBankCode, &o.PaymentAt, &o.CreatedAt, &o.UpdatedAt, &o.BuyerName)
		stats.RecentOrders = append(stats.RecentOrders, o)
	}
	if stats.RecentOrders == nil {
		stats.RecentOrders = []models.Order{}
	}
	return stats, nil
}

// FindPendingByIDPrefix finds a pending order whose UUID starts with the given prefix.
// Used by the bank webhook to match transfer notes like "PHATSHOP 4FA1B8BD" to an order.
func (r *OrderRepo) FindPendingByIDPrefix(ctx context.Context, prefix string) (*models.Order, error) {
	var o models.Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, buyer_id, total_amount, status, created_at
		FROM orders
		WHERE id::text ILIKE $1 || '%' AND status = 'pending'
		LIMIT 1`,
		prefix,
	).Scan(&o.ID, &o.BuyerID, &o.TotalAmount, &o.Status, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}
