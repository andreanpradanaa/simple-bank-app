DB_URL=postgresql://root:secret@localhost:5432/simplebank?sslmode=disable

postgres:
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simplebank

dropdb:
	docker exec -it postgres12 dropdb simplebank

migrateup:
	migrate --path db/migration --database "${DB_URL}" --verbose up 

migrateup1:
	migrate --path db/migration --database "${DB_URL}" --verbose up 1

migratedown:
	migrate --path db/migration --database "${DB_URL}" --verbose down

migratedown1:
	migrate --path db/migration --database "${DB_URL}" --verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq ${name}

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/andreanpradanaa/simple-bank-app/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/andreanpradanaa/simple-bank-app/worker TaskDistributor

protoc:
	rm -f pb/*.go
	rm -f docs/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--experimental_allow_proto3_optional \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
    proto/*.proto

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

.PHONY:
	postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock proto