%:
	@:

db-up:
	docker compose up -d

db-stop:
	docker compose stop

db-remove:
	docker compose down -v

create-migration:
	migrate create -ext sql -dir migration $(word 2, $(MAKECMDGOALS))

migrate:
	migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" up

migrate-rollback:
	migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" down 1

migrate-goto:
	migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" goto $(word 2, $(MAKECMDGOALS))

test:
	go test ./...

test-detail:
	go test -v ./...