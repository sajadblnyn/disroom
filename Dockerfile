FROM golang:1.22.2-alpine AS builder


ENV GOPROXY=https://goproxy.io,direct

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o server ./cmd/server/main.go

FROM debian:buster-slim 
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /app/server
COPY --from=builder /app/config/config.json /app/config.json

ENV CONFIG_PATH /app/

CMD [ "/app/server" ]
 