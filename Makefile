lint:
	go vet ./...

swag api:
	swag init -g cmd/api/main.go -o docs/api --exclude pkg/cosmos pkg/env