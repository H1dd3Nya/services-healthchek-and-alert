APP=services-healthchek-and-alert

build:
	go build -o $(APP) main.go

docker-build:
	docker build -t $(APP):latest .

docker-run:
	docker compose up --build

test:
	go test ./...

run:
	go run main.go 