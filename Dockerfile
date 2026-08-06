FROM golang:1.25-bookworm AS build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o dolabra -v .

FROM debian:bookworm-slim

LABEL org.opencontainers.image.source https://github.com/ocvt/dolabra
LABEL maintainer="Paul Walko <paul@bigcavemaps.com>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/* && \
    useradd --uid 1000 --create-home dolabra

# Keep the pre-multi-stage workdir so the data volume mount path is unchanged
WORKDIR /go/src/app

COPY --from=build /go/src/app/dolabra ./
COPY --from=build /go/src/app/utils/dolabra-sqlite.sql ./utils/dolabra-sqlite.sql

USER dolabra

EXPOSE 3000
CMD ["./dolabra"]
