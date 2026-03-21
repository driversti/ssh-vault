FROM golang:1-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ssh-vault ./cmd/ssh-vault

FROM alpine:3

RUN adduser -D -u 1000 vault
WORKDIR /app
COPY --from=builder /build/ssh-vault .
RUN mkdir /data && chown vault:vault /data

USER vault
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/ssh-vault", "hub", "--data", "/data"]
