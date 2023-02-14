FROM golang:1.17-buster

LABEL org.opencontainers.image.source https://github.com/ocvt/dolabra
LABEL maintainer="Paul Walko <paul@bigcavemaps.com>"

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o dolabra -v .

EXPOSE 3000
CMD ["./dolabra"]
