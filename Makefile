createdb:
	sudo docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	sudo docker exec -it postgres12 dropdb simple_bank

postgres:
	sudo docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

migrateup:
	migrate -path db/migration/ -database "postgresql://root:secret@192.168.66.16:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration/ -database "postgresql://root:secret@192.168.66.16:5432/simple_bank?sslmode=disable" -verbose up 1


migratedown:
	migrate -path db/migration/ -database "postgresql://root:secret@192.168.66.16:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration/ -database "postgresql://root:secret@192.168.66.16:5432/simple_bank?sslmode=disable" -verbose down 1


migrateupgithub:
	migrate -path db/migration/ -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedowngithub:
	migrate -path db/migration/ -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrateNew:
	migrate create -ext sql -dir db/migration -seq add_users

sqlc:
	sqlc generate

server:
	go run main.go

mock:
	mockgen -package mockdb  -destination db/mock/store.go github.com/csdengh/cur_blank/db/sqlc Store

test:
	go test -v -cover ./...

.PHONY: createdb dropdb postgres migrateup migrateup1 migratedown migratedown1 migrateNew
