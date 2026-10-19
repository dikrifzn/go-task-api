# Go Task API

RESTful Task Management API built with Go, SQLite, and Docker. This project demonstrates backend development practices, including layered architecture, authentication, database persistence, HTTP middleware, and automated testing.

## Features

- User registration and login
- Password hashing with bcrypt
- Token-based authentication
- Task CRUD operations
- Task ownership per authenticated user
- Pagination, filtering, sorting, and search
- SQLite database persistence
- HTTP request logging and request IDs
- Docker and Docker Compose support
- Unit and integration tests

## Tech Stack

- **Language:** Go
- **HTTP:** `net/http`
- **Database:** SQLite
- **Password hashing:** bcrypt
- **Containerization:** Docker and Docker Compose

## Architecture

The application uses a layered architecture:

- **Handler:** Handles HTTP requests and responses.
- **Service:** Contains business logic and validation.
- **Repository:** Handles database operations.
- **Middleware:** Handles authentication, request IDs, and logging.
- **Database:** Manages SQLite connections and migrations.

## Requirements

- Go version compatible with `go.mod`
- Docker Desktop (optional, for containerized execution)

## Run Locally

Clone the repository and enter the project directory:

```bash
git clone <YOUR_REPOSITORY_URL>
cd go-task-api
```

Download dependencies:

```bash
go mod download
```

Run the API:

```bash
go run .
```

Configure `APP_ADDR` and `DB_PATH` if needed. The default HTTP address is `:8080` when configured as described in the project.

## Run with Docker

Build and start the application:

```bash
docker compose build
docker compose up -d
```

View logs:

```bash
docker compose logs -f api
```

Stop the application:

```bash
docker compose down
```

The Compose configuration uses a named volume for database persistence. Avoid `docker compose down -v` if you want to keep the database volume.

## API Endpoints

| Method | Endpoint | Authentication |
|---|---|---|
| GET | `/health` | No |
| POST | `/register` | No |
| POST | `/login` | No |
| POST | `/tasks` | Yes |
| GET | `/tasks` | Yes |
| GET | `/tasks/{id}` | Yes |
| PUT | `/tasks/{id}` | Yes |
| DELETE | `/tasks/{id}` | Yes |

Task list query parameters may include `page`, `limit`, `completed`, `search`, and `sort`, depending on the implemented handler.

Authenticated requests use the bearer token returned by login:

```http
Authorization: Bearer YOUR_TOKEN
```

### Example: Register

```http
POST /register
Content-Type: application/json
```

Example request body (adjust to the actual `RegisterRequest` fields):

```json
{
  "name": "Example User",
  "email": "user@example.com",
  "password": "change-this-password"
}
```

### Example: Create a Task

```http
POST /tasks
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json
```

Example request body (adjust to the actual `CreateRequest` fields):

```json
{
  "title": "Learn Go testing"
}
```

## Testing

Run all tests:

```bash
go test -count=1 ./...
```

Build the application:

```bash
go build ./...
```

## Security Notes

- Passwords should be stored as bcrypt hashes, never as plaintext.
- Keep secrets and local environment files out of version control.
- Use HTTPS when exposing the API publicly.
- Configure authentication token expiry and error handling appropriately.
- Back up the SQLite database before performing destructive operations.

## Project Status

Learning and portfolio project. Verify the documented endpoints, request fields, and configuration against the current implementation before using this README as final public documentation.

## License

Choose and add an appropriate license before distributing the project.
