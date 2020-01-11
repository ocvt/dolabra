all: build

build: test
	go build -o dolabra -v

test: deps
	sqlvet .

deps:
	go get github.com/houqp/sqlvet
