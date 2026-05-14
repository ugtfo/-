.PHONY: build run test clean

build:
	go build -o log-parser ./cmd/main.go

run:
	go run ./cmd/main.go

test:
	go test ./... -v

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -f log-parser
