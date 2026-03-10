package handlers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"phatshop-backend/internal/config"
	"phatshop-backend/internal/models"
	"phatshop-backend/internal/repository"
	"phatshop-backend/internal/services"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ReceiptHandler struct {
	receipts *repository.ReceiptRepo
	orders   *repository.OrderRepo
	ocr      *services.OCRService
	cfg      *config.Config
}

func NewReceiptHandler(receipts *repository.ReceiptRepo, orders *repository.OrderRepo, ocr *services.OCRService, cfg *config.Config) *ReceiptHandler {
	return &ReceiptHandler{receipts: receipts, orders: orders, ocr: ocr, cfg: cfg}
}

// UploadReceipt handles POST /api/v1/orders/:id/receipt
// Client uploads a bank transfer receipt image; the system OCRs it,
// runs anti-fraud checks, and marks the order paid if everything passes.
func (h *ReceiptHandler) UploadReceipt(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)
	orderID := c.Param("id")

	// ── 1. Load order and verify ownership ──────────────────────────────────
	order, err := h.orders.GetByID(c.Request.Context(), orderID)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.BuyerID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if order.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "order is already " + order.Status})
		return
	}

	// ── 2. Read uploaded image (max 10 MB) ───────────────────────────────────
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20)
	file, header, err := c.Request.FormFile("receipt")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receipt image is required (field: receipt)"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpg, png, webp images are accepted"})
		return
	}

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read image"})
		return
	}

	// ── 3. ANTI-FRAUD: duplicate image hash ──────────────────────────────────
	hash := fmt.Sprintf("%x", sha256.Sum256(imgBytes))
	isDup, err := h.receipts.IsDuplicateImage(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if isDup {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "This receipt image has already been submitted. Reusing receipts is not allowed.",
			"reason": "duplicate_image",
		})
		return
	}

	// ── 4. Save image to disk ─────────────────────────────────────────────────
	dir := filepath.Join("./storage/receipts", orderID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create storage dir"})
		return
	}
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	imagePath := filepath.Join(dir, fileName)
	if err := os.WriteFile(imagePath, imgBytes, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
		return
	}

	// ── 5. OCR: extract receipt data ─────────────────────────────────────────
	if !h.ocr.IsConfigured() {
		// Clean up saved image since we can't process it
		os.Remove(imagePath)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OCR service not configured. Please ask the admin to set GEMINI_API_KEY.",
		})
		return
	}

	extracted, err := h.ocr.ExtractFromImageFile(imagePath)
	if err != nil {
		log.Printf("[receipt] OCR error for order %s: %v", orderID, err)
		os.Remove(imagePath)
		errMsg := "failed to analyse receipt image"
		if strings.Contains(err.Error(), "gemini_unavailable") {
			errMsg = "The receipt analysis service is temporarily busy. Please wait a moment and try again."
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": errMsg})
		return
	}

	// ── 6. Build verification record ─────────────────────────────────────────
	v := &models.ReceiptVerification{
		OrderID:           orderID,
		UserID:            uid,
		ImagePath:         imagePath,
		ImageHash:         hash,
		ExtractedTxnID:    extracted.TransactionID,
		ExtractedAmount:   extracted.Amount,
		ExtractedReceiver: extracted.ReceiverInfo,
		ExtractedNote:     extracted.TransferNote,
		Status:            "pending",
	}

	// ── 7. ANTI-FRAUD checks ─────────────────────────────────────────────────
	reason := h.runFraudChecks(c.Request.Context(), extracted, order)
	if reason != "" {
		v.Status = "rejected"
		v.RejectionReason = &reason
		_ = h.receipts.Create(c.Request.Context(), v)
		log.Printf("[receipt] Rejected for order %s: %s", orderID, reason)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":  reason,
			"reason": "fraud_check_failed",
			"extracted": gin.H{
				"transaction_id": extracted.TransactionID,
				"amount":         extracted.Amount,
				"receiver":       extracted.ReceiverInfo,
				"note":           extracted.TransferNote,
			},
		})
		return
	}

	// ── 8. Mark order as paid and unlock downloads ──────────────────────────
	if err := h.orders.MarkPaidByAdmin(c.Request.Context(), orderID); err != nil {
		log.Printf("[receipt] MarkPaidByAdmin error for order %s: %v", orderID, err)
		os.Remove(imagePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm order"})
		return
	}

	v.Status = "verified"
	if err := h.receipts.Create(c.Request.Context(), v); err != nil {
		// Non-fatal: order is already paid; just log
		log.Printf("[receipt] Failed to save verification record for order %s: %v", orderID, err)
	}

	log.Printf("[receipt] Order %s verified via receipt (txn: %v)", orderID, extracted.TransactionID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Payment verified. Downloads are now unlocked.",
		"extracted": gin.H{
			"transaction_id": extracted.TransactionID,
			"amount":         extracted.Amount,
			"receiver":       extracted.ReceiverInfo,
			"note":           extracted.TransferNote,
		},
	})
}

// runFraudChecks validates extracted receipt data against the order.
// Returns an empty string if everything is valid, or a human-readable rejection reason.
func (h *ReceiptHandler) runFraudChecks(ctx context.Context, ex *services.ExtractedReceipt, order *models.Order) string {
	// Check 1: suspicious image
	if ex.IsSuspicious {
		return "The receipt image appears to be edited or manipulated."
	}

	// Check 2: amount must cover order total
	if ex.Amount == nil {
		return "Could not extract the transfer amount from the receipt. Please upload a clearer image."
	}
	if *ex.Amount < order.TotalAmount {
		return fmt.Sprintf(
			"Transfer amount (%s) is less than the order total (%s).",
			formatVND(*ex.Amount), formatVND(order.TotalAmount),
		)
	}

	// Check 3: receiver must match expected account
	if ex.ReceiverInfo == nil {
		return "Could not extract the receiver information from the receipt."
	}
	receiver := strings.ToUpper(*ex.ReceiverInfo)
	expectedAcc := strings.ToUpper(h.cfg.BankAccountNo)
	expectedName := strings.ToUpper(h.cfg.BankAccountName)
	if !strings.Contains(receiver, expectedAcc) && !strings.Contains(receiver, expectedName) {
		return fmt.Sprintf(
			"Receiver on receipt (%q) does not match the expected account (%s / %s).",
			*ex.ReceiverInfo, h.cfg.BankAccountNo, h.cfg.BankAccountName,
		)
	}

	// Check 4: transaction ID must not be reused
	if ex.TransactionID != nil && *ex.TransactionID != "" {
		used, err := h.receipts.IsTxnIDUsed(ctx, *ex.TransactionID)
		if err == nil && used {
			return fmt.Sprintf("Transaction ID %q has already been used for a previous payment.", *ex.TransactionID)
		}
	}

	// Check 5: transfer note should contain the order prefix (soft warn if missing)
	orderPrefix := strings.ToUpper(order.ID[:8])
	if ex.TransferNote != nil {
		note := strings.ToUpper(*ex.TransferNote)
		if !strings.Contains(note, orderPrefix) && !strings.Contains(note, "PHATSHOP") {
			return fmt.Sprintf(
				"Transfer note (%q) does not reference this order (expected to contain %q). "+
					"Please upload the receipt for the correct order.",
				*ex.TransferNote, "PHATSHOP "+orderPrefix,
			)
		}
	}

	return ""
}

func formatVND(amount int64) string {
	return fmt.Sprintf("%d VND", amount)
}
