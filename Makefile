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

test-setup: install
	mkdir -p test-out

clean:
	rm -rf test-out main.html

test-input: test-setup
	~/go/bin/litgo --doc-out-dir test-out test/input.md

test-one-code-file: test-setup
	~/go/bin/litgo --doc-out-dir test-out test/one-code-file.md

test-two-code-files: test-setup
	~/go/bin/litgo --doc-out-dir test-out test/two-code-files.md

test-non-existent-file: test-setup
	~/go/bin/litgo --doc-out-dir test-out test/no-such-file.md

test-simple-book: test-setup
	~/go/bin/litgo --book --doc-out-dir test-out test/simple-book.md
