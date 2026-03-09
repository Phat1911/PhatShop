package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(dbURL string) (*DB, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	d.Pool.Close()
}

func (d *DB) Migrate(ctx context.Context) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE TABLE IF NOT EXISTS users (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username     VARCHAR(50)  UNIQUE NOT NULL,
			email        VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT        NOT NULL,
			display_name VARCHAR(100) DEFAULT '',
			avatar_url   TEXT         DEFAULT '',
			role         VARCHAR(20)  DEFAULT 'user',
			is_active    BOOLEAN      DEFAULT TRUE,
			created_at   TIMESTAMPTZ  DEFAULT NOW(),
			updated_at   TIMESTAMPTZ  DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name         VARCHAR(100) UNIQUE NOT NULL,
			slug         VARCHAR(100) UNIQUE NOT NULL,
			product_type VARCHAR(20)  NOT NULL,
			created_at   TIMESTAMPTZ  DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			seller_id      UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			category_id    UUID         REFERENCES categories(id) ON DELETE SET NULL,
			title          TEXT         NOT NULL,
			slug           VARCHAR(300) UNIQUE NOT NULL,
			description    TEXT         DEFAULT '',
			product_type   VARCHAR(20)  NOT NULL,
			price          BIGINT       NOT NULL CHECK (price >= 0),
			thumbnail_url  TEXT         DEFAULT '',
			preview_urls   TEXT[]       DEFAULT '{}',
			file_path      TEXT         NOT NULL,
			file_name      TEXT         NOT NULL,
			file_size      BIGINT       DEFAULT 0,
			tags           TEXT[]       DEFAULT '{}',
			is_published   BOOLEAN      DEFAULT FALSE,
			view_count     INT          DEFAULT 0,
			purchase_count INT          DEFAULT 0,
			created_at     TIMESTAMPTZ  DEFAULT NOW(),
			updated_at     TIMESTAMPTZ  DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			buyer_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			total_amount    BIGINT      NOT NULL CHECK (total_amount > 0),
			status          VARCHAR(20) DEFAULT 'pending',
			vnpay_txn_ref   VARCHAR(100) UNIQUE,
			vnpay_txn_no    VARCHAR(100),
			vnpay_bank_code VARCHAR(20),
			payment_at      TIMESTAMPTZ,
			created_at      TIMESTAMPTZ DEFAULT NOW(),
			updated_at      TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id         UUID   PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id   UUID   NOT NULL REFERENCES orders(id)   ON DELETE CASCADE,
			product_id UUID   NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			price      BIGINT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id    UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE (user_id, product_id)
		)`,
		`CREATE TABLE IF NOT EXISTS purchases (
			user_id    UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			order_id   UUID NOT NULL REFERENCES orders(id)   ON DELETE CASCADE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (user_id, product_id)
		)`,
		`CREATE TABLE IF NOT EXISTS download_tokens (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id       UUID        NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
			product_id    UUID        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			token         VARCHAR(64) UNIQUE NOT NULL,
			expires_at    TIMESTAMPTZ NOT NULL,
			used_count    INT         DEFAULT 0,
			max_uses      INT         DEFAULT 3,
			created_at    TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS receipt_verifications (
			id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id            UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			user_id             UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
			image_path          TEXT        NOT NULL,
			image_hash          TEXT        NOT NULL,
			extracted_txn_id    TEXT,
			extracted_amount    BIGINT,
			extracted_receiver  TEXT,
			extracted_note      TEXT,
			status              VARCHAR(20) DEFAULT 'pending',
			rejection_reason    TEXT,
			created_at          TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_receipt_txn_id ON receipt_verifications(extracted_txn_id) WHERE status = 'verified'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_receipt_image_hash ON receipt_verifications(image_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_receipt_order ON receipt_verifications(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_products_type      ON products(product_type)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category  ON products(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_products_published ON products(is_published)`,
		`CREATE INDEX IF NOT EXISTS idx_products_created   ON products(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_buyer       ON orders(buyer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status      ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_txn_ref     ON orders(vnpay_txn_ref)`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order  ON order_items(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cart_user          ON cart_items(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_purchases_user     ON purchases(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dl_tokens_tok      ON download_tokens(token)`,
	}

	for _, q := range queries {
		if _, err := d.Pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration error: %w\nquery: %s", err, q)
		}
	}

	for _, dir := range []string{"./uploads/thumbnails", "./uploads/previews", "./storage/products", "./storage/receipts"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Warning: could not create directory %s: %v", dir, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
