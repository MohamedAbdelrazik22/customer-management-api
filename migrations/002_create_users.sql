-- Migration: 002_create_users
-- Creates the users table for JWT authentication

CREATE TABLE IF NOT EXISTS users (
    id         INT AUTO_INCREMENT PRIMARY KEY,
    username   VARCHAR(50)  NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,          -- bcrypt hash
    created_at TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);
