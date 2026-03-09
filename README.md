# PhatShop

A Vietnamese digital goods marketplace for selling and downloading digital products (images, videos, website templates, apps). Built with a Go/Gin REST API backend and a Next.js 14 frontend.

---

## Features

- Browse and search digital products by type, keyword, price, and popularity
- Shopping cart and checkout flow
- Bank transfer payment via VietQR
- VNPay payment gateway integration
- AI-powered receipt verification using Google Gemini Vision with anti-fraud checks
- Auto-unlock downloads after payment confirmation
- Time-limited, use-counted download tokens
- Admin panel: product management, order management, user management, dashboard stats

---

## Tech Stack

### Backend
| | |
|---|---|
| Language | Go 1.22 |
| Framework | Gin |
| Database | PostgreSQL (pgx/v5) |
| Auth | JWT HS256 + bcrypt |
| OCR | Google Gemini 2.5 Flash Vision API |

### Frontend
| | |
|---|---|
| Framework | Next.js 14 (App Router, TypeScript) |
| Styling | Tailwind CSS |
| Data Fetching | TanStack React Query v5 |
| State | Zustand |
| HTTP | Axios |

---

## Project Structure

```
PhatShop/
├── start.sh                  # Starts both backend and frontend
├── backend/
│   ├── cmd/main.go           # Entry point and route registration
│   ├── internal/
│   │   ├── config/           # Environment config loader
│   │   ├── db/               # PostgreSQL connection and auto-migration
│   │   ├── models/           # Domain structs
│   │   ├── handlers/         # HTTP handlers (auth, product, order, payment, receipt, download, admin)
│   │   ├── middleware/        # JWT auth, admin role check, CORS
│   │   ├── repository/       # Raw SQL queries (pgx)
│   │   └── services/         # Gemini Vision OCR service
│   ├── uploads/              # Publicly served thumbnails and previews
│   ├── storage/              # Private product files and receipt images
│   ├── .env.example
│   └── .env
└── frontend/
    ├── app/                  # Next.js App Router pages
    │   ├── products/         # Listing and detail pages
    │   ├── cart/
    │   ├── checkout/         # Order creation and payment pages
    │   ├── orders/           # Order history and detail
    │   ├── auth/             # Login and register
    │   └── admin/            # Admin dashboard, products, orders, users
    ├── components/           # Navbar, ProductCard, DownloadButton, etc.
    └── lib/                  # Axios instance, Zustand stores
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+ and npm
- PostgreSQL running on `localhost:5432`

### 1. Database Setup

```sql
CREATE USER phatshop WITH PASSWORD 'phatshop123';
CREATE DATABASE phatshop OWNER phatshop;
```

> Tables are created automatically on first backend run.

### 2. Backend Setup

```bash
cd backend
cp .env.example .env
# Edit .env and fill in required values (see Environment Variables below)
go run ./cmd/main.go
```

Backend starts on `http://localhost:8080`. API docs: see [API Endpoints](#api-endpoints).

### 3. Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

Frontend starts on `http://localhost:3000`.

### 4. One-Command Start (after first setup)

```bash
./start.sh
```

Starts the backend in the background and the frontend in the foreground. Press `Ctrl+C` to stop both.

---

## Environment Variables

Copy `backend/.env.example` to `backend/.env` and fill in the values.

| Variable | Required | Description |
|---|---|---|
| `PORT` | No | Server port (default: `8080`) |
| `DB_URL` | Yes | PostgreSQL DSN, e.g. `postgres://phatshop:phatshop123@localhost:5432/phatshop?sslmode=disable` |
| `JWT_SECRET` | Yes | Secret key for signing JWT tokens |
| `JWT_EXPIRES_IN` | No | Token lifetime (default: `168h`) |
| `UPLOAD_DIR` | No | Public uploads directory (default: `./uploads`) |
| `STORAGE_DIR` | No | Private file storage directory (default: `./storage`) |
| `MAX_UPLOAD_SIZE` | No | Max upload size in bytes (default: `209715200` = 200MB) |
| `GEMINI_API_KEY` | Yes* | Google AI Studio API key for receipt OCR |
| `BANK_ACCOUNT_NO` | Yes* | Your bank account number for anti-fraud validation |
| `BANK_ACCOUNT_NAME` | Yes* | Your bank account name for anti-fraud validation |
| `VNPAY_TMN_CODE` | Yes* | VNPay terminal code |
| `VNPAY_HASH_SECRET` | Yes* | VNPay hash secret |
| `VNPAY_URL` | No | VNPay payment URL (default: sandbox) |
| `FRONTEND_URL` | No | Frontend URL for VNPay return redirect |

> \* Required for full payment functionality. Basic browsing and cart work without these.

---

## Payment Methods

### 1. Bank Transfer via VietQR
A QR code is displayed on the payment page. The buyer scans it and transfers the exact amount with the transfer note shown (e.g. `PHATSHOP ABC123`). The order can be confirmed via:
- **Receipt OCR** (buyer uploads a screenshot of their bank receipt — see below)
- **Admin manual confirmation** via admin panel

### 2. Receipt OCR Verification
The buyer uploads a screenshot of their bank transfer receipt. The system:
1. Sends the image to **Google Gemini 2.5 Flash Vision** to extract: transaction ID, amount, receiver info, and transfer note
2. Runs **anti-fraud checks**:
   - Duplicate image hash detection
   - Suspicious/edited image flag from AI
   - Amount must match order total
   - Receiver account must match configured bank account
   - Transaction ID must not have been used before
   - Transfer note must contain the order prefix
3. On success, order is marked paid and downloads are unlocked immediately

### 3. VNPay
Standard Vietnamese payment gateway. Buyer is redirected to VNPay's checkout and returns to the site after payment. Confirmation is received via VNPay IPN callback.

---

## API Endpoints

### Public
```
GET  /api/health
GET  /api/v1/categories
GET  /api/v1/products                    ?type=&sort=&search=&page=&limit=
GET  /api/v1/products/:id
GET  /api/v1/downloads/file?token=       File download (token-based, no JWT)
GET  /api/v1/payments/vnpay/ipn          VNPay IPN callback
```

### Auth
```
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Protected (JWT required)
```
GET  /api/v1/users/me
PUT  /api/v1/users/me

GET    /api/v1/cart
POST   /api/v1/cart
DELETE /api/v1/cart/:product_id
DELETE /api/v1/cart

POST /api/v1/orders
GET  /api/v1/orders
GET  /api/v1/orders/:id
POST /api/v1/orders/:id/receipt          Upload receipt image for OCR verification

POST /api/v1/payments/vnpay/create

GET  /api/v1/downloads/request/:product_id   Generate download token
GET  /api/v1/downloads/check/:product_id     Check purchase status
```

### Admin (JWT + admin role)
```
GET    /api/v1/admin/stats
GET    /api/v1/admin/products
POST   /api/v1/admin/products
DELETE /api/v1/admin/products/:id
PATCH  /api/v1/admin/products/:id/publish
GET    /api/v1/admin/categories
POST   /api/v1/admin/categories
PUT    /api/v1/admin/categories/:id
DELETE /api/v1/admin/categories/:id
GET    /api/v1/admin/orders
GET    /api/v1/admin/orders/:id
PATCH  /api/v1/admin/orders/:id/status
GET    /api/v1/admin/users
PATCH  /api/v1/admin/users/:id/role
```

---

## Default URLs

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080/api/v1 |
| Health check | http://localhost:8080/api/health |
| Uploaded files | http://localhost:8080/uploads/ |
