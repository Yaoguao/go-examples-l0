FROM golang:1.24-bullseye AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash gcc make git libc6-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/app ./cmd/wb-examples-l0

RUN go build -o /app/bin/prod ./cmd/producer

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

#                 STAGE 2
FROM debian:bullseye-slim AS app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libssl1.1 libc6 libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/app /app
COPY --from=builder /app/config /config

CMD ["/app"]

#                 STAGE 3
FROM debian:bullseye-slim AS producer

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libssl1.1 libc6 libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/prod /prod
COPY --from=builder /app/config /config

CMD ["/prod"]

#                BUILD MIGRATOR

FROM debian:bullseye-slim AS migrator

RUN apt-get update && apt-get install -y --no-install-recommends postgresql-client

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations /app/migrations

ENTRYPOINT ["/bin/sh", "-c", "migrate -path=/app/migrations -database=postgres://wbexaml0db:wbexam@db/wbexaml0db?sslmode=disable up"]
