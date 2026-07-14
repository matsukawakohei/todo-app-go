.PHONY: db-up db-stop db-remove create-migration migrate migrate-rollback migrate-goto test test-detail action-lint lint fmt fmt-f

DB_URL := postgres://todo:todo@localhost:5432/todo?sslmode=disable
TEST_DB_URL := postgres://todo:todo@localhost:5432/todo_test?sslmode=disable

%:
	@:

app:
	GO_ENV=dev go run main.go

db-up:
	docker compose up -d --wait

db-stop:
	docker compose stop

db-remove:
	docker compose down -v

create-migration:
ifeq ($(word 2, $(MAKECMDGOALS)),)
	$(error マイグレーションファイル名を指定してください。例: make create-migration add_users_table)
	
endif
	migrate create -ext sql -dir migration $(word 2, $(MAKECMDGOALS))

migrate:
	migrate -source file://migration -database "$(DB_URL)" up

migrate-rollback:
	migrate -source file://migration -database "$(DB_URL)" down 1

migrate-goto:
ifeq ($(word 2, $(MAKECMDGOALS)),)
	$(error マイグレーションのバージョンを指定してください。例: make migrate-goto 20260625170155)
endif
	migrate -source file://migration -database "$(DB_URL)" goto $(word 2, $(MAKECMDGOALS))

test:
	TEST_DB_URL=$(TEST_DB_URL) go test ./... -count=1

test-detail:
	TEST_DB_URL=$(TEST_DB_URL) go test -v ./... -count=1

lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.12.2 golangci-lint run

fmt:
	gofmt -l .

fmt-f:
	gofmt -w .

action-lint:
	docker run --rm -v "$(shell pwd):$(shell pwd)" -w "$(shell pwd)" rhysd/actionlint:latest