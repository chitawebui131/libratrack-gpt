ARG GO_VERSION=1.25

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/api ./cmd/api && \
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/seed ./cmd/seed && \
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/import ./cmd/import
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/seed /app/seed
COPY --from=builder /out/import /app/import
COPY migrations /app/migrations
ENV MIGRATIONS_PATH=/app/migrations
EXPOSE 8080
USER app
CMD ["/app/api"]
