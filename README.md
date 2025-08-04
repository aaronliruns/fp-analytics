# FP Analytics

A fingerprint analytics service built with Go and Gin framework that collects and stores device fingerprints with deduplication capabilities.

## Features

- **RESTful API** for fingerprint collection
- **SQLite3 database** with indexed key column for optimal performance
- **Deduplication logic** prevents duplicate fingerprints based on unique keys
- **Docker support** for easy deployment
- **Volume mounting** for persistent data storage

## Database Schema

The application uses SQLite3 with the following schema:

```sql
CREATE TABLE fingerprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    filename TEXT NOT NULL
);

CREATE INDEX idx_fingerprints_key ON fingerprints(key);
```

## API Endpoints

### Collect Fingerprint

**Endpoint:** `POST /v1/finger/collect/{key}`

- **key**: Unique identifier passed as URL parameter (not in JSON payload)
- **Body**: JSON payload containing fingerprint data

**Success Response:**
- **Status:** 201 Created
- **Body:** `{"filename": "generated_filename.enc", "key": "your-key", "duplicate": false}`

**Duplicate Response:**
- **Status:** 409 Conflict  
- **Body:** `{"message": "Fingerprint with this key already exists", "duplicate": true}`

**Error Responses:**
- **400 Bad Request:** Missing key parameter or invalid JSON
- **500 Internal Server Error:** Database or file system errors

## Building and Running

### Prerequisites

- Go 1.23 or later (for local development)
- Docker (for containerized deployment)

### Local Development

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd fp-analytics
   go mod download
   ```

2. **Set environment variables:**
   ```bash
   export PROFILE_PATH=/path/to/profiles
   ```

3. **Run the application:**
   ```bash
   go run main.go app.go
   ```

### Docker Deployment

1. **Create host directory:**
   ```bash
   mkdir -p /tmp/fingerprints
   ```

2. **Build Docker image:**
   ```bash
   docker build -t fp-analytics .
   ```

3. **Run with Docker:**
   ```bash
   docker run -d \
     --name fp-analytics-server \
     -p 8090:8090 \
     -v /root/fingerprints:/app/profiles \
     -e PROFILE_PATH=/app/profiles \
     fp-analytics
   ```

4. **Using Docker Compose (if available):**
   ```bash
   docker-compose up -d
   ```

## Testing the API

### Basic Test

```bash
curl -X POST http://localhost:8090/v1/finger/collect/test-key-123 \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "device-abc-123",
    "browser": "Chrome",
    "version": "120.0.0.0",
    "os": "Linux",
    "screen_resolution": "1920x1080",
    "timezone": "UTC+8",
    "language": "en-US",
    "user_agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
    "fingerprint_data": {
      "canvas": "canvas_hash_value_here",
      "webgl": "webgl_renderer_info",
      "fonts": ["Arial", "Helvetica", "Times New Roman"],
      "plugins": ["Chrome PDF Plugin", "Native Client"]
    },
    "timestamp": "2025-08-02T15:39:58+08:00"
  }'
```

### Test Deduplication

1. **First request (should return 201 Created):**
   ```bash
   curl -X POST http://localhost:8090/v1/finger/collect/unique-key-001 \
     -H "Content-Type: application/json" \
     -d '{"test": "data", "value": 123}'
   ```

2. **Duplicate request (should return 409 Conflict):**
   ```bash
   curl -X POST http://localhost:8090/v1/finger/collect/unique-key-001 \
     -H "Content-Type: application/json" \
     -d '{"different": "payload", "but_same_key": true}'
   ```

### Test Error Handling

```bash
# Missing key parameter (should return 400 Bad Request)
curl -X POST http://localhost:8090/v1/finger/collect/ \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

## Configuration

The application uses `config.yaml` for configuration:

```yaml
server:
  port: "8090"

fingerprints:
  profile_path: ${PROFILE_PATH}
  version: 138
```

## Data Storage

- **Database:** SQLite3 file created in the same directory as profiles
- **Fingerprint files:** Stored as encrypted `.enc` files with generated filenames
- **Filename format:** `{hash}_VERSION_{version}_{date}.enc`
- **Deduplication:** Based on unique key parameter, not JSON payload content

## Recent Changes

- **v2.0:** Refactored endpoint to use key as URL parameter instead of JSON payload field
- **Performance:** Added database index on key column for faster lookups
- **Docker:** Added comprehensive Docker support with volume mounting
- **API:** Improved error handling and response consistency