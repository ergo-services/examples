FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY . .

RUN go mod download

ARG NODE_CMD=node1
RUN CGO_ENABLED=0 GOOS=linux go build -tags latency,pprof -o /app ./cmd/${NODE_CMD}

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app .

CMD ["./app"]
