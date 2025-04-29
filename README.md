---

# **Finly Backend**

**Finly** is a backend application designed to help users manage their personal finances. It enables efficient tracking of expenses and income, budget management, and custom categories for enhanced financial organization. Built using **Go**, the backend provides a **RESTful API** for seamless frontend integration.

## 🚀 Features

- **User Authentication**: Register, login, logout, refresh token, and fetch user profile.
- **Budget Management**: Create budgets, view budget details, check balances, and view transaction history.
- **Transaction Management**: Add, update, delete, and list transactions (deposits/withdrawals).
- **Category Management**: Create, retrieve, and delete custom transaction categories.
- **Secure API**: JWT-based authentication for securing endpoints.
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

## 📝 Setup

To run the application locally, follow these steps:

1. **Create a `.env` file**:  
   In the root of the project, create a `.env` file with the following content:

    ```env
    HTTP_PORT=8080

    DB_USERNAME=your_db_username
    DB_PASSWORD=your_db_password
    DB_HOST=localhost
    DB_PORT=5432
    DB_NAME=finly
    DB_SSLMODE=disable

    REDIS_HOST=localhost
    REDIS_PORT=6379
    REDIS_PASSWORD=your_redis_password
    REDIS_DB=0
    ```

   - **ENV**: Set this to `dev` for local development. In production, use `prod`.
   - **HTTP_PORT**: Define the port for the HTTP server to listen on (default is `8080`).
   - **Database**: Configure your PostgreSQL database credentials (`DB_USERNAME`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_SSLMODE`).
   - **Redis**: Set up Redis credentials (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`).

2. **Install Dependencies**:  
   Make sure you have PostgreSQL and Redis installed or use Docker to run them in containers.

3. **Run Database Migrations**:  
   Before starting the application, apply the database migrations to set up the necessary database schema:

    ```bash
    make migrate-up
    ```

4. **Launch the Application**:  
   After setting up the `.env` file and running the migrations, you can start the application and the necessary services (PostgreSQL and Redis) in containers by running the following command:

    ```bash
    make up
    ```
