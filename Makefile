DB_URL=postgresql://root:secret@postgres:5432/simple_bank?sslmode=disable
GO=go
BIN_DIR=build
BIN_NAME=main
GOOS=linux
GOARCH=amd64
# To connect to psql in docker-compose
# 1) docker inspect b031546e59ac44c42f9ddd52ef1160d7f475d50e936758fc600be36a9fc382dd | grep IPAddress
# 2) docker compose run postgres bash 
# 3) psql -h 172.19.0.3 -U root -d simple_bank

local_run: docker_build
	docker-compose up --force-recreate

compile:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(BIN_DIR)/$(BIN_NAME) main.go

docker_build: compile
	docker build -t simple-bank-api .

network:
	docker network create bank-network

postgres:
	docker run --name postgres --network bank-network -p 49152:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret postgres:14-alpine

redis:
	docker run --name redis -p 6380:6379 redis:7-alpine

mysql:
	docker run --name mysql8 -p 3306:3306  -e MYSQL_ROOT_PASSWORD=secret -d mysql:8

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	docker compose run api migrate -path ./migration -database "$(DB_URL)" -verbose up

migrateup1:
	docker compose run api migrate -path ./migration -database "$(DB_URL)" -verbose up 1

migratedown:
	docker compose run api migrate -path ./migration -database "$(DB_URL)" -verbose down

migratedown1:
	docker compose run api migrate -path ./migration -database "$(DB_URL)" -verbose down 1

new_migration:
	docker compose run api migrate create -ext sql -dir db/migration -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination ./db/mock/store.go github.com/40grivenprog/simple-bank/db/sqlc Store
	mockgen -package mockwk -destination ./worker/mock/distributor.go github.com/40grivenprog/simple-bank/worker TaskDistributor

proto:
	rm -f pb/*.proto
	rm -f doc/swagger/*.swagger.json
	protoc --experimental_allow_proto3_optional --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
	proto/*.proto
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: compile local_run docker_build network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis
