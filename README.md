# Customer Management API

A REST API for managing customers, built with **Go**, **Gin**, and **MySQL**.

---

## Technologies

| Technology | Purpose |
|---|---|
| Go | Programming language |
| Gin | HTTP web framework |
| MySQL | Relational database |
| database/sql | Standard Go database interface |
| godotenv | Load environment variables from `.env` |
| golang-jwt/jwt | JWT authentication |
| Git / GitHub | Version control |

---

## Requirements

Make sure you have the following installed before running the project:

- [Go 1.21+](https://go.dev/dl/)
- [MySQL 8.0+](https://dev.mysql.com/downloads/)
- Git

---

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/MohamedAbdelrazik22/customer-management-api.git
cd customer-management-api
```

### 2. Create the MySQL database

Log in to MySQL and create the database:

```sql
CREATE DATABASE customer_management;
```

### 3. Run the SQL migrations

```bash
mysql -u root -p customer_management < migrations/001_create_customers.sql
mysql -u root -p customer_management < migrations/002_create_users.sql
```

### 4. Configure environment variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

Edit `.env`:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=customer_management
SERVER_PORT=8080
JWT_SECRET=your-secure-secret
```

### 5. Install Go dependencies

```bash
go mod download
```

### 6. Run the application

```bash
go run main.go
```

The server will start at `http://localhost:8080`.

---

## JWT Configuration

The application uses JWT (JSON Web Tokens) for authentication.

```env
JWT_SECRET=your-secure-secret
```

> **The application will not start if `JWT_SECRET` is missing or empty.**

- The secret must be set as an environment variable (or in your `.env` file).
- There is no default or fallback secret — this is intentional to prevent insecure deployments.
- Use a long, random string in production (e.g., at least 32 characters).

---

## Authentication

Some endpoints require authentication. To access protected endpoints, first obtain a JWT token by logging in, then include it in the `Authorization` header of your requests.

### Login

```http
POST /login
Content-Type: application/json
```

**Request body:**

```json
{
  "username": "admin",
  "password": "yourpassword"
}
```

**Response `200 OK`:**

```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Using the Token

Include the token in subsequent requests to protected endpoints:

```http
Authorization: Bearer <token>
```

### Protected Endpoints

| Method | Endpoint | Auth Required |
|---|---|---|
| GET | `/customers` | No |
| GET | `/customers/:id` | No |
| POST | `/customers` | **Yes** |
| PUT | `/customers/:id` | **Yes** |
| DELETE | `/customers/:id` | **Yes** |

---

## API Endpoints

### Get All Customers

```http
GET /customers
```

**Query Parameters (optional):**
- `page` — page number (default: 1)
- `limit` — items per page (default: 10, max: 100)
- `search` — search by name or email

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": 1,
      "name": "Mohamed Abdelrazik",
      "email": "Mohamed@example.com",
      "status": "active",
      "created_at": "2026-09-11T10:30:00Z"
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 1,
  "total_pages": 1
}
```

---

### Get Customer by ID

```http
GET /customers/:id
```

**Response `200 OK`:**

```json
{
  "data": {
    "id": 1,
    "name": "Mohamed Abdelrazik",
    "email": "Mohamed@example.com",
    "status": "active",
    "created_at": "2026-09-11T10:30:00Z"
  }
}
```

**Response `400 Bad Request`** — invalid (non-numeric) customer ID:

```json
{
  "error": "Invalid customer ID"
}
```

**Response `404 Not Found`:**

```json
{
  "error": "Customer not found"
}
```

---

### Create Customer

```http
POST /customers
Authorization: Bearer <token>
Content-Type: application/json
```

**Request body:**

```json
{
  "name": "Mohamed Abdelrazik",
  "email": "Mohamed@example.com",
  "status": "active"
}
```

**Validation rules:**

- `name` — required, max 100 characters
- `email` — required, valid format, must be unique
- `status` — required, must be `"active"` or `"inactive"`

**Response `201 Created`:**

```json
{
  "data": {
    "id": 1,
    "name": "Mohamed Abdelrazik",
    "email": "Mohamed@example.com",
    "status": "active",
    "created_at": "2026-09-11T10:30:00Z"
  }
}
```

**Response `409 Conflict`** — duplicate email:

```json
{
  "error": "Email already exists"
}
```

**Response `401 Unauthorized`** — missing or invalid token:

```json
{
  "error": "Authorization header is required"
}
```

---

### Update Customer

```http
PUT /customers/:id
Authorization: Bearer <token>
Content-Type: application/json
```

**Request body:**

```json
{
  "name": "Mohamed Mohamed",
  "email": "Mohamed@example.com",
  "status": "inactive"
}
```

**Response `200 OK`** — returns the updated customer.

**Response `404 Not Found`** — customer does not exist.

**Response `409 Conflict`** — email is already used by another customer.

---

### Delete Customer

```http
DELETE /customers/:id
Authorization: Bearer <token>
```

**Response `200 OK`:**

```json
{
  "message": "Customer deleted successfully"
}
```

**Response `404 Not Found`:**

```json
{
  "error": "Customer not found"
}
```

---

## Duplicate Email Protection

Customer email addresses are **unique**. Attempting to create a customer with an email address already in use will return:

```http
409 Conflict
```

```json
{
  "error": "Email already exists"
}
```

The uniqueness is enforced at both the application level and the database level (via a `UNIQUE` constraint on the `email` column).

---

## HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request (validation error or invalid ID) |
| 401 | Unauthorized (missing or invalid JWT token) |
| 404 | Not Found |
| 409 | Conflict (duplicate email) |
| 500 | Internal Server Error |

---

## Running Tests

```bash
go test ./...
```

Tests cover:

- JWT configuration validation (missing / empty `JWT_SECRET`)
- Duplicate customer email returns `409 Conflict`
- Invalid customer ID returns `400 Bad Request`
- Unauthorized requests return `401 Unauthorized`
- Customer CRUD operations
- Input validation rules

---

## Project Structure

```
customer-management-api/
│
├── main.go                          # Entry point, startup validation, graceful shutdown
│
├── config/
│   └── database.go                  # MySQL connection setup
│
├── models/
│   ├── customer.go                  # Customer struct and input types
│   └── user.go                      # User struct and login input
│
├── handlers/
│   ├── auth_handler.go              # Login endpoint and JWT token generation
│   └── customer_handler.go          # HTTP request handling & validation
│
├── middleware/
│   └── auth.go                      # JWT authentication middleware
│
├── repositories/
│   ├── customer_repository.go       # SQL queries & database operations (context-aware)
│   └── user_repository.go           # User lookup for authentication (context-aware)
│
├── routes/
│   └── routes.go                    # Route registration
│
├── migrations/
│   ├── 001_create_customers.sql     # Customers table schema (with UNIQUE email)
│   └── 002_create_users.sql         # Users table schema
│
├── main_test_helpers/
│   └── config.go                    # Shared config validation (testable)
│
├── tests/
│   └── config_test.go               # JWT config validation tests
│
├── .env.example                     # Environment variable template
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

**Request flow:**

```
Routes → Middleware (Auth) → Handlers → Repositories → MySQL
```

---

## License

This project is open-source and available for educational use.
