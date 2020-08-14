THIS_FILE := $(lastword $(MAKEFILE_LIST))

all: build test

full-test: clean start test
	@$(MAKE) -f $(THIS_FILE) clean

test:
	sleep 1
	python3 tests/main.py

start:
	./launch.sh

build: format static-check
	docker build -t ocvt/dolabra:latest .

format:
	gofmt -w .

static-check: deps
	sqlvet .
	staticcheck ./app ./app/handler ./utils

deps:
	go get github.com/houqp/sqlvet
	go get honnef.co/go/tools/cmd/staticcheck

clean:
	rm -f dolabra
	(docker stop dolabra && docker rm dolabra) || true
	rm -f data/dolabra-sqlite.sqlite3
