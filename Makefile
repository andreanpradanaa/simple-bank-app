postgres:
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simplebank

dropdb:
	docker exec -it postgres12 dropdb simplebank

migrateup:
	migrate --path db/migration --database "postgres://root:password@simple-bank.cnyocsq4m6kx.us-east-2.rds.amazonaws.com:5432/simple_bank" --verbose up 

migrateup1:
	migrate --path db/migration --database "postgres://root:secret@localhost:5432/simplebank?sslmode=disable" --verbose up 1

migratedown:
	migrate --path db/migration --database "postgresql://root:secret@localhost:5432/simplebank?sslmode=disable" --verbose down

migratedown1:
	migrate --path db/migration --database "postgresql://root:secret@localhost:5432/simplebank?sslmode=disable" --verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/andreanpradanaa/simple-bank-app/db/sqlc Store

.PHONY:
	postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock