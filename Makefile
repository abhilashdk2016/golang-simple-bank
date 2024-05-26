migration:
	@migrate create -ext sql -dir db/migration -seq $(filter-out $@, $(MAKECMDGOALS))
migrateup:
	migrate -path db/migration -database "postgresql://postgres:abhi@2024@localhost:5432/golang-bank?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgresql://postgres:abhi@2024@localhost:5432/golang-bank?sslmode=disable" -verbose down
migrateup1:
	migrate -path db/migration -database "postgresql://postgres:abhi@2024@localhost:5432/golang-bank?sslmode=disable" -verbose up 1
migratedown1:
	migrate -path db/migration -database "postgresql://postgres:abhi@2024@localhost:5432/golang-bank?sslmode=disable" -verbose down 1
sqlc:
	sqlc generate
test:
	@go test -v -cover -short ./...
server:
	@go run main.go
generatemock:
	mockgen -package mockdb -destination db/mock/store.go github.com/abhilashdk2016/golang-simple-bank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/abhilashdk2016/golang-simple-bank/worker TaskDistributor
genproto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative proto/*.proto --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative
evans:
	evans --host localhost --port 8990 -r repl
