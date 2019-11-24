FROM golang:1.13-alpine

LABEL maintainer="Paul Walko <paul@seaturtle.pw"

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go build -o dolabra .

EXPOSE 3000

CMD ["./dolabra"]
