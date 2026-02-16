FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN apk add --no-cache make

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app && \
    make build-agent

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/downloads ./downloads

EXPOSE 8080
CMD ["./main"]
