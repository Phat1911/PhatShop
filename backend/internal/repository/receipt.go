package repository

import (
	"context"
	"errors"
	"phatshop-backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceiptRepo struct {
	pool *pgxpool.Pool
}

func NewReceiptRepo(pool *pgxpool.Pool) *ReceiptRepo {
	return &ReceiptRepo{pool: pool}
}

func (r *ReceiptRepo) Create(ctx context.Context, v *models.ReceiptVerification) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO receipt_verifications
			(order_id, user_id, image_path, image_hash, extracted_txn_id, extracted_amount, extracted_receiver, extracted_note, status, rejection_reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at`,
		v.OrderID, v.UserID, v.ImagePath, v.ImageHash,
		v.ExtractedTxnID, v.ExtractedAmount, v.ExtractedReceiver, v.ExtractedNote,
		v.Status, v.RejectionReason,
	).Scan(&v.ID, &v.CreatedAt)
}

// IsDuplicateImage returns true if this exact image hash has been submitted before.
func (r *ReceiptRepo) IsDuplicateImage(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM receipt_verifications WHERE image_hash=$1)`, hash,
	).Scan(&exists)
	return exists, err
}

// IsTxnIDUsed returns true if this transaction ID has already been used for a verified receipt.
func (r *ReceiptRepo) IsTxnIDUsed(ctx context.Context, txnID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM receipt_verifications WHERE extracted_txn_id=$1 AND status='verified')`, txnID,
	).Scan(&exists)
	return exists, err
}

// FindByOrderID returns the most recent receipt submission for an order.
func (r *ReceiptRepo) FindByOrderID(ctx context.Context, orderID string) (*models.ReceiptVerification, error) {
	var v models.ReceiptVerification
	err := r.pool.QueryRow(ctx, `
		SELECT id, order_id, user_id, image_path, image_hash,
		       extracted_txn_id, extracted_amount, extracted_receiver, extracted_note,
		       status, rejection_reason, created_at
		FROM receipt_verifications WHERE order_id=$1 ORDER BY created_at DESC LIMIT 1`, orderID,
	).Scan(&v.ID, &v.OrderID, &v.UserID, &v.ImagePath, &v.ImageHash,
		&v.ExtractedTxnID, &v.ExtractedAmount, &v.ExtractedReceiver, &v.ExtractedNote,
		&v.Status, &v.RejectionReason, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}
