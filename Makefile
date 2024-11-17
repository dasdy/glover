.PHONY: test
test:
	mkdir -p ./tmp
	go test -race  -coverprofile=./tmp/c.out ./...

.PHONY: benchmark
benchmark:
	go test ./... -bench=. -run=^#

# .PHONY: profile
# profile:
# 	go test ./db/... -bench=. -benchtime=5s -cpuprofile ./tmp/cpu.prof -run=^#
# 	go tool pprof -http=:8080 ./tmp/cpu.prof

.PHONY: cover
cover: test
	 go tool cover -html=./tmp/c.out -o coverage.html


.PHONY: build
build:
	mkdir -p ./tmp
	go build -v -race -o ./tmp/glover cmd/main.go

.PHONY: clean
clean:
	rm -rvf tmp

.PHONY: run-dev
run-dev:
	make -j3 templates-watch tailwind-watch server-watch

.PHONY: server-watch
server-watch:
	air

.PHONY: templates
templates:
	templ generate web/components

.PHONY: templates-watch
templates-watch:
	templ generate --watch web/components

.PHONY: tailwind
tailwind:
	tailwindcss -i assets/css/tailwind_input.css -o assets/css/tailwind_output.css

.PHONY: tailwind-watch
tailwind-watch:
	tailwindcss -i assets/css/tailwind_input.css -o assets/css/tailwind_output.css --watch

.PHONY: lint
lint:
	golangci-lint run
	prettier . --write
