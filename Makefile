ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(ALL_PACKAGES)

tidy:
	go mod tidy

serve: fmt vet
	env $(cat dev.env | xargs) go run cmd/*.go

build: fmt vet
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/openwebui-telegram ./cmd
