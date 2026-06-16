FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ARG TARGET=server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app ./cmd/${TARGET}

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/app /app/app
USER app
CMD ["./app"]
