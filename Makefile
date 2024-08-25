build:
	go build -o bin/go_crypto_exchange

run: build
	./bin/go_crypto_exchange

test:
	go test -v ./...
