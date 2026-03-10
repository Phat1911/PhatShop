package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DBURL           string
	JWTSecret       string
	JWTExpires      string
	UploadDir       string
	StorageDir      string
	MaxUploadSize   int64
	VNPayTmnCode    string
	VNPayHashSecret string
	VNPayURL        string
	FrontendURL     string
	SePayAPIKey     string // SePay webhook API key for verifying incoming webhooks
	GeminiAPIKey    string // Google Gemini API key for receipt OCR
	BankAccountNo   string // Expected receiver bank account number
	BankAccountName string // Expected receiver account name
	AdminEmail      string // Seed admin email
	AdminUsername   string // Seed admin username
	AdminPassword   string // Seed admin password
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "209715200"), 10, 64)

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DBURL:           getEnv("DB_URL", "postgres://phatshop:phatshop123@localhost:5432/phatshop?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		JWTExpires:      getEnv("JWT_EXPIRES_IN", "168h"),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		StorageDir:      getEnv("STORAGE_DIR", "./storage"),
		MaxUploadSize:   maxUpload,
		VNPayTmnCode:    getEnv("VNPAY_TMN_CODE", ""),
		VNPayHashSecret: getEnv("VNPAY_HASH_SECRET", ""),
		VNPayURL:        getEnv("VNPAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		SePayAPIKey:     getEnv("SEPAY_API_KEY", ""),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		BankAccountNo:   getEnv("BANK_ACCOUNT_NO", ""),
		BankAccountName: getEnv("BANK_ACCOUNT_NAME", ""),
		AdminEmail:      getEnv("ADMIN_EMAIL", ""),
		AdminUsername:   getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:   getEnv("ADMIN_PASSWORD", ""),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set in .env or environment")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
