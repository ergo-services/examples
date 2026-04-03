FROM golang:1.25-alpine AS deps

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM deps AS build-node1
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/node1

FROM deps AS build-node2
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/node2

FROM deps AS build-node3
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/node3

FROM deps AS build-node4
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/node4

FROM deps AS build-node5
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/node5

FROM deps AS build-observer
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/observer

FROM deps AS build-slowweb
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/slowweb

FROM alpine:3.21 AS base
RUN apk --no-cache add ca-certificates
WORKDIR /root/

FROM base AS node1
COPY --from=build-node1 /app ./app
CMD ["./app"]

FROM base AS node2
COPY --from=build-node2 /app ./app
CMD ["./app"]

FROM base AS node3
COPY --from=build-node3 /app ./app
CMD ["./app"]

FROM base AS node4
COPY --from=build-node4 /app ./app
CMD ["./app"]

FROM base AS node5
COPY --from=build-node5 /app ./app
CMD ["./app"]

FROM base AS observer
COPY --from=build-observer /app ./app
CMD ["./app"]

FROM alpine:3.21 AS slowweb
WORKDIR /root/
COPY --from=build-slowweb /app ./app
CMD ["./app"]
