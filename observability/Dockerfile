FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/
COPY apps/ ./apps/

ARG NODE_CMD=node1
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o app ./cmd/${NODE_CMD}

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/app .

CMD ["./app"]
