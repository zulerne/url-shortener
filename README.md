# URL Shortener

A robust, production-ready URL Shortener service built with Go (Golang).
It features a clean architecture, Docker support, comprehensive testing (Unit & E2E), and CI/CD integration.

## 🚀 Features

- **Shorten URLs**: Create short aliases for long URLs.
- **Redirection**: Fast redirection (307 Temporary Redirect) to the original URL.
- **Custom Aliases**: User can specify a custom alias or let the service generate a random one.
- **Authentication**: Usage is protected via Basic Auth.
- **Persistent Storage**: Utilizes SQLite for data persistence.
- **Dockerized**: Fully containerized for easy development and deployment.
- **Tests**: Covered by Unit tests and E2E functional tests.

## 🛠 Tech Stack

- **Language**: Go 1.25+
- **Database**: SQLite3
- **Router**: Standard `net/http` ServeMux
- **Validation**: `go-playground/validator`
- **Logging**: `slog` (Structured Logging)

## 🏁 Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose (optional, for containerized run)
- Make (optional, for using Makefile commands)

### 1. Local Development

Clone the repository and install dependencies:

```bash
git clone https://github.com/yourusername/url-shortener.git
cd url-shortener
go mod download
```

**Running the application:**

You can run the application directly on your machine. It will use a local `storage.db` file.

```bash
make run
```
_Or manually: `go run cmd/url-shortener/main.go`_

The server will start at `http://localhost:8080` (default).

### 2. Testing

The project includes both Unit and Integration (E2E) tests.

- **Run Unit Tests:**
  ```bash
  make test
  ```

- **Run E2E Tests:**
  This launches the application in a separate process and tests it as a black box using `httpexpect`.
  ```bash
  make test-e2e
  ```

### 3. Running with Docker (Local)

To run the application in a production-like environment using Docker:

```bash
# 1. Create .env file
cp .env.example .env

# 2. Build and Start
make docker-up
```

The API will be available at `http://localhost:8080`.

To stop the container:
```bash
make docker-down
```

## 📦 Deployment

This project supports a modern containerized deployment workflow.

### 1. Build and Publish Image

To deploy updates, first build and push the Docker image to your registry (e.g., Docker Hub).
Ensure you are logged in (`docker login`).

```bash
export DOCKER_USERNAME={username}
make publish
```
_This will build `{username}/url-shortener:latest` and push it to Docker Hub._

### 2. Deploy on Server

On your production server:

1.  Copy `docker-compose.prod.yml` and rename it to `docker-compose.yml`.
2.  Create a `.env` file with your production secrets.
3.  Run the application:

```bash
# Pull the latest image and start
export DOCKER_IMAGE={docker-image}
docker compose pull && docker compose up -d
```

## 🔌 API Reference

**Auth**: Basic Auth is required for creates. Default: `admin` / `admin`.

### 1. Create Short URL

**POST** `/url`

**Request Body:**
```json
{
  "url": "https://google.com",
  "alias": "google"  // Optional. If omitted, random alias is generated.
}
```

**Response (200 OK):**
```json
{
  "status": "OK",
  "alias": "google"
}
```

### 2. Redirect

**GET** `/{alias}`

**Response:**
- `307 Temporary Redirect` to the original URL.
- `404 Not Found` if alias does not exist.

## 📂 Project Structure

```
.
├── cmd/                # Main applications
│   └── url-shortener   # Entry point
├── internal/           # Private application logic
│   ├── config/         # Configuration loading
│   ├── server/         # HTTP server and handlers
│   │   ├── handler/    # API handlers & business logic
│   │   └── middleware/ # HTTP middlewares (Auth, Logger, etc)
│   ├── storage/        # Storage interfaces & implementation (SQLite)
│   └── lib/            # Shared utilities
├── tests/              # End-to-End tests
├── Dockerfile          # Multi-stage build definition
├── docker-compose.yml  # Local dev environment
└── Makefile            # Make commands
```
