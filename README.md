# CityNext Appointment Booking API

A REST service for booking appointments at the CityNext office, built with Go, [Huma framework](https://huma.rocks/), and SQLite.

## Features

- POST `/appointments` for booking appointments
- **Validation Rules**:
  - Prevents appointment scheduling on weekends
  - Prevents booking on UK public holidays (via Nager.Date API)
  - Prevents booking dates in the past
  - Prevents duplicate appointments per date
- Repository pattern with interfaces for easy testing
- SQLite with GORM 
- Unit and integration tests

## Project Structure

```
citynext/
├── cmd/server/                 # Main application entry point
├── internal/
│   ├── api/                    # API layer (handlers, models, routes)
│   ├── database/               # Database layer (models, repositories)
│   ├── services/               # Business logic layer
│   └── config/                 # Configuration management
├── pkg/client/                 # External API clients
└── tests/                      # Test files
    ├── unit/                   # Unit tests
    └── integration/            # Integration tests
```

## Getting Started

### Prerequisites

- Go 1.22 or later

### Running the application:
```bash
go run cmd/server/main.go
```

The server will start on port 9119 by default.

### Configuration

The application can be configured using environment variables:

- `SERVER_PORT`: Server port (default: 9119)
- `DB_PATH`: SQLite database file path (default: citynext.db)

Example:
```bash
export SERVER_PORT=8080
export DB_PATH=/path/to/citynext.db
go run cmd/server/main.go
```

## API Documentation

Once the server is running, you can access the interactive API documentation at:
- http://localhost:9119/docs

### Endpoint

#### POST /appointments

Creates a new appointment.

**Request Body:**
```json
{
  "firstName": "John",
  "lastName": "Doe",
  "visitDate": "2025-09-25"
}
```

**Response:**
```json
{
  "id": 1,
  "firstName": "John",
  "lastName": "Doe",
  "visitDate": "2025-09-25",
  "createdAt": "2025-07-04T10:30:00Z"
}
```

**Validation Rules:**
- `firstName` and `lastName` are required and must not be empty
- `visitDate` must be in the future
- `visitDate` must not be a UK public holiday
- `visitDate` must not fall on a weekend
- Only one appointment per date is allowed

**Error Responses:**
- `422 Unprocessable Entity`: Validation errors
- `500 Internal Server Error`: Server errors

## Testing

### Running Tests

Run all tests:
```bash
go test ./...
```

Run unit tests only:
```bash
go test ./tests/unit/...
```

Run integration tests only:
```bash
go test ./tests/integration/...
```
