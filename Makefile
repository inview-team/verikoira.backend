CMD := verikoira

build:
	go build -o $(CMD) ./cmd/main.go

run:
	make build
	./$(CMD)

lint:
	golangci-lint run ./...

.PHONY: build run lint
