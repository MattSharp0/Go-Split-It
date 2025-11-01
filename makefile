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


# start docker container for DB
postgres: 
	docker run --name postgres17a -p 5432:5432 -e POSTGRES_USER=$(DATABASE_USER) -e POSTGRES_PASSWORD=$(DATABASE_PASSWORD) -d postgres:17-alpine

# create txsplitdb database
createdb:
    docker exec -it postgres17a createdb --username=root --owner=root $(DATABASE_NAME)

# drop txsplitdb
dropdb:
	docker exec -it postgres17a dropdb $(DATABASE_NAME)

migrateup:
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose up 1

migratedown:
	migrate -path db/migrations -database "$(DATABASE_URL)" -verbose down 1

sqlc:
	sqlc generate
    
# docker_build:
#     docker build --build-arg DATABASE_URL="$(DATABASE_URL)" -t splitapp .

.PHONY: hello postgres createdb dropdb sqlc migrateup migratedown