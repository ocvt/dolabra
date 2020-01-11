all: build

integration-test: build
	./launch.sh
	# TODO run python tests
	# TODO clean up

build: format static-check
	docker build -t ocvt/dolabra:latest .

format:
	go fmt

static-check: deps
	go vet
	sqlvet .

deps:
	go get github.com/houqp/sqlvet
