build: tangle
	go build

tangle:
	clear
	./bin/lit *.lit
	go fmt ./...
