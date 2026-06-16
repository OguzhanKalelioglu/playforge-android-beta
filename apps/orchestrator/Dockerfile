FROM golang:1.25-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates android-tools

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/orchestrator ./cmd/orchestrator

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata android-tools && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/orchestrator /app/orchestrator

USER app
WORKDIR /app

CMD ["./orchestrator"]
