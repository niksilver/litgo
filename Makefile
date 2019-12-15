build: tangle
	go build

tangle:
	clear
	./bin/lit main.lit
	go fmt ./...

test-input: tangle
	~/go/bin/litgo test/input.md

test-one-code-file: tangle
	~/go/bin/litgo test/one-code-file.md

test-two-code-files: tangle
	~/go/bin/litgo test/two-code-files.md

