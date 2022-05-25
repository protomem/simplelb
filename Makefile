
.PHONY: build
build:
	go build -o ./build/ ./cmd/lb

.PHONY: run
run: build
	./build/lb


