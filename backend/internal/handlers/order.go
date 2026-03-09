package handlers

import (
	"net/http"
	"phatshop-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orders *repository.OrderRepo
	cart   *repository.CartRepo
}

func NewOrderHandler(orders *repository.OrderRepo, cart *repository.CartRepo) *OrderHandler {
	return &OrderHandler{orders: orders, cart: cart}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(string)

	items, _, err := h.cart.GetItems(c.Request.Context(), uid)
	if err != nil || len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}

	order, err := h.orders.Create(c.Request.Context(), uid, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orders, err := h.orders.ListByUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")
	id := c.Param("id")
	order, err := h.orders.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.BuyerID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	c.JSON(http.StatusOK, order)
}
