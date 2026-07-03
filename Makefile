.PHONY: db-up db-stop db-remove create-migration migrate migrate-rollback migrate-goto test test-detail

DB_URL := postgres://todo:todo@localhost:5432/todo?sslmode=disable

%:
	@:

db-up:
	docker compose up -d

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
	go test ./...

test-detail:
	go test -v ./...