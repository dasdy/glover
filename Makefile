.PHONY: test
test:
	mkdir -p ./tmp
	go test -race -v -coverprofile=./tmp/c.out ./...

.PHONY: cover
cover: test
	 go tool cover -html=./tmp/c.out -o coverage.html


.PHONY: build
build:
	go build -race -o bin/glover main.go

.PHONY: clean
clean:
	rm -rvf bin tmp

.PHONY: run-dev
run-dev:
	air

.PHONY: tailwind
tailwind:
	tailwindcss -i css/input.css -o assets/css/output.css

.PHONY: tailwind-watch
tailwind-watch:
	tailwindcss -i css/input.css -o assets/css/output.css --watch

.PHONY: lint
lint:
	golangci-lint run
	go vet
	prettier . --write
