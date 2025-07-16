up:
	docker compose up -d

down:
	docker compose down -v

test:
	go test ./...

run:
	go run ./cmd/main.go
