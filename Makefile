CMD := verikoira

deploy:
	docker-compose -f ./deployment/docker-compose.yml up --build

build:
	go build -o $(CMD) ./cmd/main.go

run:
	make build
	./$(CMD)

lint:
	golangci-lint run ./...

.PHONY: build run lint deploy
