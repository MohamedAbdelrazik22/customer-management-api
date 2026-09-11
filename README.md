# Customer Management API

A simple REST API for managing customers, built with **Go**, **Gin**, and **MySQL**.

---

## Technologies

| Technology | Purpose |
|---|---|
| Go | Programming language |
| Gin | HTTP web framework |
| MySQL | Relational database |
| database/sql | Standard Go database interface |
| godotenv | Load environment variables from `.env` |
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

### 3. Run the SQL migration

```bash
mysql -u root -p customer_management < migrations/001_create_customers.sql
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

## API Endpoints

### Get All Customers

```http
GET /customers
```

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
  ]
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

---

### Update Customer

```http
PUT /customers/:id
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

## HTTP Status Codes

| Code | Meaning |
|---|---|
| 200 | OK |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 404 | Not Found |
| 409 | Conflict (duplicate email) |
| 500 | Internal Server Error |

---

## Project Structure

```
customer-management-api/
│
├── main.go                          # Entry point
│
├── config/
│   └── database.go                  # MySQL connection setup
│
├── models/
│   └── customer.go                  # Customer struct and input types
│
├── handlers/
│   └── customer_handler.go          # HTTP request handling & validation
│
├── repositories/
│   └── customer_repository.go       # SQL queries & database operations
│
├── routes/
│   └── routes.go                    # Route registration
│
├── migrations/
│   └── 001_create_customers.sql     # Database schema
│
├── .env.example                     # Environment variable template
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

**Request flow:**

```
Routes → Handlers → Repositories → MySQL
```

---

## License

This project is open-source and available for educational use.
