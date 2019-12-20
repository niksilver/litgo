build: tangle
	go build

test: tangle
	go test

tangle:
	clear
	./bin/lit main.lit
	go fmt ./...

install: tangle
	go install

test-input: install
	~/go/bin/litgo test/input.md

test-one-code-file: install
	~/go/bin/litgo test/one-code-file.md

test-two-code-files: install
	~/go/bin/litgo test/two-code-files.md

test-non-existent-file: install
	~/go/bin/litgo test/no-such-file.md

test-simple-book: install
	~/go/bin/litgo --book test/simple-book.md
