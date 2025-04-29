include .env
export $(shell sed 's/=.*//g' .env)

up:
	ENV=dev go run cmd/server/main.go

migrate-up:
	goose -dir migrations postgres "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	goose -dir migrations postgres "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" reset

swagger:
	swag init --parseDependency --parseInternal -g cmd/server/main.go

dbuild:
	docker build -t my-go-worker . && docker run --rm -p 8080:8080 my-go-worker

lint:
	 golangci-lint run --config .golangci.yml

test:
	set -o pipefail && go test ./... -json | tparse -all