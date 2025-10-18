.PHONY: proto build run-auth run-gateway run-video run-local docker-up docker-down clean tidy

proto:
	cd shared && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

tidy:
	cd shared && go mod tidy
	cd gateway && go mod tidy
	cd auth-service && go mod tidy
	cd video-service && go mod tidy

build:
	cd gateway && go build -o ../bin/gateway ./cmd/main.go
	cd auth-service && go build -o ../bin/auth-service ./cmd/main.go
	cd video-service && go build -o ../bin/video-service ./cmd/main.go

run-auth:
	cd auth-service && go run ./cmd/main.go

run-gateway:
	cd gateway && go run ./cmd/main.go

run-video:
	cd video-service && go run ./cmd/main.go

run-local:
	@echo "Starting auth service..."
	cd auth-service && go run ./cmd/main.go & 
	@sleep 2
	@echo "Starting video service..."
	cd video-service && go run ./cmd/main.go &
	@sleep 2
	@echo "Starting gateway..."
	cd gateway && go run ./cmd/main.go

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

clean:
	rm -rf bin/
	docker-compose down -v

test:
	cd shared && go test ./...
	cd gateway && go test ./...
	cd auth-service && go test ./...
	cd video-service && go test ./...
