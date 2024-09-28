hello:
	echo "Hello"

build:
	go build -o ./api-run cmd/main.go --debug

run:
	go run cmd/main.go --debug