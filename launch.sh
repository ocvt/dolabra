#!/bin/bash

set -e

up () {
  docker run \
    --name dolabra \
    --detach \
    --restart unless-stopped \
    --env-file dolabra.env \
    --volume $PWD/data:/go/src/app/data:rw \
    --publish 3000:3000 \
    ocvt/dolabra:latest

  docker system prune -af
}

down () {
  stop
  docker rm -f dolabra || true
  rm -f data/dolabra-sqlite.sqlite3
}

stop () {
  docker stop dolabra || true
}

logs () {
  docker logs -f dolabra
}

###

build () {
  format
  static-check
  docker build -t ocvt/dolabra:latest .
}

deps () {
	go get github.com/houqp/sqlvet
	go get honnef.co/go/tools/cmd/staticcheck
}

format () {
  gofmt -w .
}

full-test () {
  down
  up
  test
  down
}

static-check () {
  sqlvet .
  staticcheck ./app ./app/handler ./utils
}

test () {
  sleep 1
  python3 tests/main.py
}


$@
