lint:
	go vet ./...

swag api:
	swag init -g internal/app/api/api.go -o docs/api --exclude pkg/cosmos pkg/env