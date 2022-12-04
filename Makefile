lint:
	go vet ./...

swag gateway:
	swag init -g cmd/gateway/main.go -o cmd/gateway/docs --exclude pkg