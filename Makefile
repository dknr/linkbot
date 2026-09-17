.PHONY: build run vet clean

BIN := linkbot

build:
	go build -tags goolm -o $(BIN) .

run: build
	./$(BIN) -config config.json

vet:
	go vet -tags goolm ./...

clean:
	rm -f $(BIN)
