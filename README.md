# **Finly Backend**

**Finly** is a backend application for managing personal finances. It enables users to track expenses and income, manage budgets, and create custom categories for better financial organization. The backend is built with **Go** and offers a **RESTful API** for seamless integration with frontend applications.

---

## 📚 Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [API Documentation](#api-documentation)
- [Key Endpoints](#key-endpoints)
    - [Authentication](#authentication)
    - [Budget](#budget)
    - [Category](#category)
    - [Transaction](#transaction)

---

## 🚀 Features

- **User Authentication**: Register, login, logout, refresh token, and fetch user profile.
- **Budget Management**: Create budgets, view details, check balances, and transaction history.
- **Transaction Management**: Add, update, delete, and list transactions (deposits/withdrawals).
- **Category Management**: Create, retrieve, and delete custom transaction categories.
- **Secure API**: JWT-based authentication for protecting endpoints.
- **Persistent Data Storage**: PostgreSQL for storage and Redis for caching.

---

## 🛠️ Technologies Used

- **Language**: Go (v1.23.4)
- **Framework**: Echo (v4.13.3)
- **Database**: PostgreSQL
- **Caching**: Redis
- **Authentication**: JWT (JSON Web Tokens)
- **Logging**: Zap
- **API Documentation**: Swagger

---

## 📖 API Documentation

Access the Swagger UI after starting the server at:  
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## 📌 Key Endpoints

### 🔐 Authentication

- `POST /auth/register` – Register a new user
- `POST /auth/login` – Authenticate a user and receive a JWT token
- `POST /auth/refresh` – Refresh the JWT token
- `POST /auth/logout` – Invalidate the user's token
- `POST /auth/me` – Retrieve the authenticated user's profile

---

### 💰 Budget

- `POST /budget` – Create a new budget
- `GET /budget/{budget_id}` – Retrieve budget details by ID
- `GET /budget/{budget_id}/balance` – Get the current balance
- `GET /budget/{budget_id}/history` – View budget transaction history

---

### 🗂️ Category

- `POST /category` – Create a new custom category
- `GET /category` – List all categories for the user
- `GET /category/{id}` – Retrieve a category by ID
- `DELETE /category/{id}` – Delete a category

---

### 💸 Transaction

- `POST /transaction` – Create a new transaction (deposit or withdrawal)
- `GET /transaction` – List all transactions
- `PUT /transaction/{id}` – Update a transaction
- `DELETE /transaction/{id}` – Delete a transaction
