# Load dev environment variables
include .dev.env
export $(shell sed 's/=.*//' .dev.env)

# test variables are loaded correctly
hello: 
	@echo "Makefile executed"
	@echo "Database URL: " $(DATABASE_URL)
	@echo "Database NAME: " $(DATABASE_NAME)
	@echo "Database USER: " $(DATABASE_USER)
	@echo "Database PW: " $(DATABASE_PASSWORD)
	@echo "Environment: " $(ENVIROMENT)
	@echo "DB Container Name: " $(DB_CONTAINER_NAME)


# start docker container for DB
postgres: 
	docker run --name $(DB_CONTAINER_NAME) -p 5432:5432 -e POSTGRES_USER=$(DATABASE_USER) -e POSTGRES_PASSWORD=$(DATABASE_PASSWORD) -d postgres:postgres:18.0-alpine3.22

# create txsplitdb database
createdb:
    docker exec -it $(DB_CONTAINER_NAME) createdb --username=root --owner=root $(DATABASE_NAME)

# drop txsplitdb
dropdb:
	docker exec -it $(DB_CONTAINER_NAME) dropdb $(DATABASE_NAME)

migrateup:
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose up 1

migratedown:
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test ./... -v

test-coverage:
	go test ./... -cover

test-coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"
    
# docker_build:
#     docker build --build-arg DATABASE_URL="$(DATABASE_URL)" -t splitapp .

.PHONY: hello postgres createdb dropdb sqlc migrateup migratedown test test-coverage test-coverage-html