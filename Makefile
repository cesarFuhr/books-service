up inmemory:
	docker compose -f docker-compose.inmemory.yml up -d --build

down inmemory:
	docker compose -f docker-compose.inmemory.yml down

up postgres:
	docker compose -f docker-compose.postgres.yml up -d --build

down postgres:
	docker compose -f docker-compose.postgres.yml down

test: test-env-up
	DATABASE_URL='postgres://root:root@localhost:5432/booksdb?sslmode=disable' \
	DATABASE_MIGRATIONS_PATH='../../../migrations' \
	go test -p=1 ./...

test-env-up:
	docker compose -f docker-compose.test.yml up -d

test-env-down:
	docker compose -f docker-compose.test.yml down