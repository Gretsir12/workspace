build:
	go mod tidy
	go build -o bin/backend ./cmd/
run:
	go mod tidy
	go build -o bin/backend ./cmd/
	./backend
