all: build integration-test

full-test: clean integration-test clean

integration-test:
	./launch.sh
	sleep 1
	python3 tests/main.py

build: format static-check
	docker build -t ocvt/dolabra:latest .

format:
	gofmt -w .

static-check: deps
	go vet
	sqlvet .

deps:
	go get github.com/houqp/sqlvet

clean:
	rm -f dolabra
	(docker stop dolabra && docker rm dolabra) || true
	rm -f data/dolabra-sqlite.sqlite3
