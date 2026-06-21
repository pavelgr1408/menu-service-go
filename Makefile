.PHONY: test build run compose-up compose-down compose-logs clean

test:
	go test ./...

build:
	go build -trimpath -o bin/menu-service ./cmd/menu-service

run:
	go run ./cmd/menu-service

compose-up:
	docker compose up --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f restaurant-menu-service restaurant-menu-postgres

clean:
	rm -rf bin
