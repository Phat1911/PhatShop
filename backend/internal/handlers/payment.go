package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"phatshop-backend/internal/config"
	"phatshop-backend/internal/repository"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	orders *repository.OrderRepo
	cfg    *config.Config
}

func NewPaymentHandler(orders *repository.OrderRepo, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{orders: orders, cfg: cfg}
}

func hmacSHA512(key, data string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func buildVNPayURL(baseURL string, params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var rawParts []string
	for _, k := range keys {
		rawParts = append(rawParts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	rawQuery := strings.Join(rawParts, "&")
	secureHash := hmacSHA512(secret, rawQuery)
	return baseURL + "?" + rawQuery + "&vnp_SecureHash=" + secureHash
}

func (h *PaymentHandler) CreatePaymentURL(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orders.GetByID(c.Request.Context(), req.OrderID)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.BuyerID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order is not pending"})
		return
	}

	txnRef := fmt.Sprintf("%s-%s", order.ID[:8], time.Now().Format("20060102150405"))
	if err := h.orders.UpdateTxnRef(c.Request.Context(), order.ID, txnRef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
		return
	}

	params := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_Command":    "pay",
		"vnp_TmnCode":    h.cfg.VNPayTmnCode,
		"vnp_Amount":     fmt.Sprintf("%d", order.TotalAmount*100),
		"vnp_CreateDate": time.Now().Format("20060102150405"),
		"vnp_CurrCode":   "VND",
		"vnp_IpAddr":     c.ClientIP(),
		"vnp_Locale":     "vn",
		"vnp_OrderInfo":  fmt.Sprintf("Thanh toan don hang %s", order.ID[:8]),
		"vnp_OrderType":  "other",
		"vnp_ReturnUrl":  h.cfg.FrontendURL + "/payment/return",
		"vnp_TxnRef":     txnRef,
	}

	paymentURL := buildVNPayURL(h.cfg.VNPayURL, params, h.cfg.VNPayHashSecret)
	c.JSON(http.StatusOK, gin.H{"payment_url": paymentURL})
}

func (h *PaymentHandler) VNPayIPN(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if k != "vnp_SecureHash" && k != "vnp_SecureHashType" {
			params[k] = v[0]
		}
	}
	receivedHash := c.Query("vnp_SecureHash")

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var rawParts []string
	for _, k := range keys {
		rawParts = append(rawParts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	rawQuery := strings.Join(rawParts, "&")
	computedHash := hmacSHA512(h.cfg.VNPayHashSecret, rawQuery)

	if !hmac.Equal([]byte(computedHash), []byte(receivedHash)) {
		c.JSON(http.StatusOK, gin.H{"RspCode": "97", "Message": "Invalid signature"})
		return
	}

	responseCode := c.Query("vnp_ResponseCode")
	txnRef := c.Query("vnp_TxnRef")
	txnNo := c.Query("vnp_TransactionNo")
	bankCode := c.Query("vnp_BankCode")

	if responseCode != "00" {
		c.JSON(http.StatusOK, gin.H{"RspCode": "00", "Message": "Confirm success"})
		return
	}

	if err := h.orders.MarkPaid(c.Request.Context(), txnRef, txnNo, bankCode); err != nil {
		c.JSON(http.StatusOK, gin.H{"RspCode": "99", "Message": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"RspCode": "00", "Message": "Confirm success"})
}
