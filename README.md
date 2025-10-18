# Text-Tube Microservices

A microservices-based application built with Go Workspaces, featuring API Gateway, Authentication Service, Video Service, gRPC communication, and MongoDB.

## Architecture

- **Gateway Service**: HTTP/JSON API gateway (Port 8080)
- **Auth Service**: gRPC-based authentication service (Port 50051)
- **Video Service**: gRPC-based YouTube video service (Port 50052)
- **MongoDB**: Data persistence (Port 27017)
- **Shared Module**: Common protobuf definitions

## Features

- 🔐 JWT-based authentication
- 🎥 YouTube channel and video search
- 💾 MongoDB caching (30-minute cache)
- 🚀 gRPC inter-service communication
- 🐳 Docker containerization
- 📦 Go Workspaces for unified development

## Project Structure

```
text-tube/
├── go.work                    # Go workspace file
├── shared/                    # Shared module (protobuf definitions)
│   ├── go.mod
│   └── proto/
│       ├── auth.proto
│       └── video.proto
├── gateway/                   # API Gateway service
│   ├── go.mod
│   ├── cmd/
│   └── internal/
│       ├── handler/
│       ├── client/
│       └── middleware/
├── auth-service/              # Authentication service
│   ├── go.mod
│   ├── cmd/
│   └── internal/
│       ├── service/
│       ├── repository/
│       └── models/
├── video-service/             # Video service
│   ├── go.mod
│   ├── cmd/
│   └── internal/
│       ├── service/
│       ├── repository/
│       ├── models/
│       └── client/
├── docker-compose.yml
└── Makefile
```

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Protocol Buffers compiler (protoc)
- YouTube Data API v3 Key ([Get one here](https://console.cloud.google.com/apis/credentials))

## Quick Start

### 1. Setup Environment

Copy the example environment file and add your YouTube API key:

```bash
cp .env.example .env
# Edit .env and add your YOUTUBE_API_KEY
```

### 2. Generate Protocol Buffers

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.32.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
make proto
```

### 3. Tidy Dependencies

```bash
make tidy
```

### 4. Run with Docker

```bash
# Set your YouTube API key
export YOUTUBE_API_KEY=your_api_key_here

# Start all services
make docker-up

# View logs
make docker-logs
```

### 5. Run Locally (for development)

```bash
# Start MongoDB
docker run -d -p 27017:27017 --name text-tube-mongo mongo:7

# Set environment variables
export YOUTUBE_API_KEY=your_api_key_here
export JWT_SECRET=your-secret-key

# Run all services
make run-local

# Or run services separately in different terminals:
make run-auth      # Terminal 1
make run-video     # Terminal 2
make run-gateway   # Terminal 3
```

## API Endpoints

### Authentication

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secret123"
  }'
```

Save the token from the response!

#### Get Profile (Protected)
```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Video Operations (All Protected)

#### Search Channel and Get Videos
```bash
curl "http://localhost:8080/api/videos/search?channel=TechChannel" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:
```json
{
  "channel_id": "UC...",
  "channel_title": "Tech Channel",
  "channel_description": "...",
  "thumbnail_url": "https://...",
  "videos": [
    {
      "video_id": "abc123",
      "title": "Video Title",
      "description": "...",
      "thumbnail_url": "https://...",
      "published_at": "2024-01-01T00:00:00Z",
      "channel_id": "UC...",
      "channel_title": "Tech Channel"
    }
  ]
}
```

#### Get Channel Videos by ID
```bash
curl "http://localhost:8080/api/videos/channel/UC_CHANNEL_ID?max_results=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Get Video Details
```bash
curl "http://localhost:8080/api/videos/VIDEO_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response includes view count and like count:
```json
{
  "video": {
    "video_id": "abc123",
    "title": "Video Title",
    "description": "...",
    "thumbnail_url": "https://...",
    "published_at": "2024-01-01T00:00:00Z",
    "channel_id": "UC...",
    "channel_title": "Tech Channel",
    "view_count": 1000000,
    "like_count": 50000
  }
}
```

## Complete Test Script

```bash
#!/bin/bash

# Register
echo "=== Registering User ==="
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"pass123"}'

echo -e "\n\n=== Logging In ==="
# Login and save token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"pass123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo "Token: $TOKEN"

echo -e "\n\n=== Searching for Channel ==="
curl "http://localhost:8080/api/videos/search?channel=Fireship" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n=== Getting Video Details ==="
curl "http://localhost:8080/api/videos/VIDEO_ID" \
  -H "Authorization: Bearer $TOKEN"
```

## Caching

The video service implements intelligent caching:

- **Channel data**: Cached for 30 minutes
- **Video metadata**: Cached for 30 minutes
- **Benefits**: Reduces YouTube API quota usage and improves response times

## Environment Variables

### Gateway
- `GATEWAY_PORT`: Gateway port (default: 8080)
- `AUTH_SERVICE_ADDR`: Auth service address (default: localhost:50051)
- `VIDEO_SERVICE_ADDR`: Video service address (default: localhost:50052)

### Auth Service
- `AUTH_SERVICE_PORT`: Auth service port (default: 50051)
- `MONGO_URI`: MongoDB connection string (default: mongodb://localhost:27017)
- `JWT_SECRET`: JWT signing secret

### Video Service
- `VIDEO_SERVICE_PORT`: Video service port (default: 50052)
- `MONGO_URI`: MongoDB connection string (default: mongodb://localhost:27017)
- `YOUTUBE_API_KEY`: **Required** - Your YouTube Data API v3 key

## Development Commands

```bash
# Generate protobuf code
make proto

# Tidy all module dependencies
make tidy

# Build binaries
make build

# Run services locally
make run-local

# Run with Docker
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down

# Clean everything
make clean

# Run tests
make test
```

## YouTube API Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable "YouTube Data API v3"
4. Create credentials (API Key)
5. Copy the API key and add it to your `.env` file

**Note**: YouTube Data API has quota limits. The free tier provides 10,000 units per day. Each search costs 100 units, and each video list costs 1 unit.

## MongoDB Collections

### `users`
Stores user authentication data

### `channels`
Caches YouTube channel information

### `videos`
Caches YouTube video metadata

## Security Notes

⚠️ **Important for Production**:

1. Change the `JWT_SECRET` to a strong, random value
2. Keep your `YOUTUBE_API_KEY` private
3. Use environment variables, never commit secrets
4. Consider implementing rate limiting
5. Add HTTPS/TLS for production

## Stop Services

```bash
make docker-down
```

## Troubleshooting

### "YouTube API error: 403"
- Check if your API key is valid
- Verify YouTube Data API v3 is enabled in your Google Cloud project
- Check if you've exceeded your quota

### "Failed to connect to MongoDB"
- Ensure MongoDB is running
- Check the `MONGO_URI` environment variable

### "Invalid token"
- Token may have expired (24-hour expiry)
- Login again to get a new token

## License

MIT
