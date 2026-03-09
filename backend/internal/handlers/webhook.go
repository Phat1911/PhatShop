package handlers

import (
	"log"
	"net/http"
	"phatshop-backend/internal/config"
	"phatshop-backend/internal/repository"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	orders *repository.OrderRepo
	cfg    *config.Config
}

func NewWebhookHandler(orders *repository.OrderRepo, cfg *config.Config) *WebhookHandler {
	return &WebhookHandler{orders: orders, cfg: cfg}
}

// SePayWebhook receives transaction notifications from SePay.
// SePay docs: https://my.sepay.vn/userapi/transactions/list
// Payload sent by SePay when a transfer is received:
//
//	{
//	  "id": 12345,
//	  "gateway": "VPBank",
//	  "transactionDate": "2024-01-01 12:00:00",
//	  "accountNumber": "0764717493",
//	  "code": null,
//	  "content": "PHATSHOP 4FA1B8BD",
//	  "transferType": "in",
//	  "transferAmount": 50000,
//	  "accumulated": 100000,
//	  "subAccount": null,
//	  "referenceCode": "FT2400112345",
//	  "description": "PHATSHOP 4FA1B8BD"
//	}
func (h *WebhookHandler) SePayWebhook(c *gin.Context) {
	// Verify API key sent by SePay in the "apikey" header
	if h.cfg.SePayAPIKey != "" {
		apiKey := c.GetHeader("apikey")
		if apiKey != h.cfg.SePayAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
	}

	var payload struct {
		ID             int64   `json:"id"`
		Gateway        string  `json:"gateway"`
		AccountNumber  string  `json:"accountNumber"`
		Content        string  `json:"content"`
		TransferType   string  `json:"transferType"`
		TransferAmount float64 `json:"transferAmount"`
		ReferenceCode  string  `json:"referenceCode"`
		Description    string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Only process incoming transfers
	if payload.TransferType != "in" {
		c.JSON(http.StatusOK, gin.H{"message": "ignored"})
		return
	}

	// Extract order prefix from transfer content/description
	// Expected format: "PHATSHOP 4FA1B8BD ..." or "PHATSHOP4FA1B8BD"
	note := strings.ToUpper(payload.Content)
	if note == "" {
		note = strings.ToUpper(payload.Description)
	}

	orderPrefix := extractOrderPrefix(note)
	if orderPrefix == "" {
		log.Printf("[webhook] No PHATSHOP order prefix found in note: %q", note)
		c.JSON(http.StatusOK, gin.H{"message": "no matching order prefix"})
		return
	}

	log.Printf("[webhook] SePay: amount=%.0f, note=%q, order_prefix=%s", payload.TransferAmount, note, orderPrefix)

	// Find the pending order
	order, err := h.orders.FindPendingByIDPrefix(c.Request.Context(), orderPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if order == nil {
		log.Printf("[webhook] No pending order found for prefix: %s", orderPrefix)
		c.JSON(http.StatusOK, gin.H{"message": "order not found or already paid"})
		return
	}

	// Verify transfer amount covers order total
	if int64(payload.TransferAmount) < order.TotalAmount {
		log.Printf("[webhook] Amount mismatch: sent %.0f, need %d for order %s",
			payload.TransferAmount, order.TotalAmount, order.ID)
		c.JSON(http.StatusOK, gin.H{"message": "amount too low"})
		return
	}

	// Mark order as paid and unlock downloads
	if err := h.orders.MarkPaidByAdmin(c.Request.Context(), order.ID); err != nil {
		log.Printf("[webhook] MarkPaidByAdmin error for order %s: %v", order.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm order"})
		return
	}

	log.Printf("[webhook] Order %s confirmed paid via SePay (ref: %s)", order.ID, payload.ReferenceCode)
	c.JSON(http.StatusOK, gin.H{"message": "order confirmed", "order_id": order.ID})
}

// extractOrderPrefix pulls the 8-char order ID from a transfer note.
// Matches patterns like "PHATSHOP 4FA1B8BD" or "PHATSHOP4FA1B8BD"
var noteRe = regexp.MustCompile(`PHATSHOP\s*([0-9A-F]{8})`)

func extractOrderPrefix(note string) string {
	m := noteRe.FindStringSubmatch(strings.ToUpper(note))
	if len(m) < 2 {
		return ""
	}
	return strings.ToLower(m[1])
}
