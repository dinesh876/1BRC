.PHONY: all
all: build run

.PHONY: build
build:
	@go build -o bin/1brc

.PHONY: run 
run:
	@bin/1brc
