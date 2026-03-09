package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	ProductType string    `json:"product_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID            string    `json:"id"`
	SellerID      string    `json:"seller_id"`
	CategoryID    *string   `json:"category_id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	ProductType   string    `json:"product_type"`
	Price         int64     `json:"price"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	PreviewURLs   []string  `json:"preview_urls"`
	FilePath      string    `json:"-"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	Tags          []string  `json:"tags"`
	IsPublished   bool      `json:"is_published"`
	ViewCount     int       `json:"view_count"`
	PurchaseCount int       `json:"purchase_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	SellerName    string    `json:"seller_name,omitempty"`
	CategoryName  string    `json:"category_name,omitempty"`
}

type Order struct {
	ID            string      `json:"id"`
	BuyerID       string      `json:"buyer_id"`
	TotalAmount   int64       `json:"total_amount"`
	Status        string      `json:"status"`
	VNPayTxnRef   *string     `json:"vnpay_txn_ref"`
	VNPayTxnNo    *string     `json:"vnpay_txn_no"`
	VNPayBankCode *string     `json:"vnpay_bank_code"`
	PaymentAt     *time.Time  `json:"payment_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Items         []OrderItem `json:"items,omitempty"`
	BuyerName     string      `json:"buyer_name,omitempty"`
}

type OrderItem struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	ProductID string    `json:"product_id"`
	Price     int64     `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	Product   *Product  `json:"product,omitempty"`
}

type CartItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
	Product   *Product  `json:"product,omitempty"`
}

type Purchase struct {
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	OrderID   string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

type DownloadToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UsedCount int       `json:"used_count"`
	MaxUses   int       `json:"max_uses"`
	CreatedAt time.Time `json:"created_at"`
}

type ReceiptVerification struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	UserID            string    `json:"user_id"`
	ImagePath         string    `json:"-"`
	ImageHash         string    `json:"-"`
	ExtractedTxnID    *string   `json:"extracted_txn_id"`
	ExtractedAmount   *int64    `json:"extracted_amount"`
	ExtractedReceiver *string   `json:"extracted_receiver"`
	ExtractedNote     *string   `json:"extracted_note"`
	Status            string    `json:"status"`
	RejectionReason   *string   `json:"rejection_reason"`
	CreatedAt         time.Time `json:"created_at"`
}

// Request/Response types

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	ProductType string `json:"product_type" binding:"required,oneof=image video website_app"`
}

type AdminStats struct {
	TotalProducts int64   `json:"total_products"`
	TotalOrders   int64   `json:"total_orders"`
	TotalRevenue  int64   `json:"total_revenue"`
	TotalUsers    int64   `json:"total_users"`
	RecentOrders  []Order `json:"recent_orders"`
}
