all: build

build: format test
	go build -o dolabra -v

format:
	go fmt

test: deps
	go vet
	sqlvet .

deps:
	go get github.com/houqp/sqlvet
