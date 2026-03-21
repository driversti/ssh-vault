FROM golang:1-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the hub binary for the container platform
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ssh-vault ./cmd/ssh-vault

# Cross-compile client binaries for all supported platforms
RUN CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -ldflags="-s -w" -o dist/ssh-vault_linux_amd64  ./cmd/ssh-vault && \
    CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build -ldflags="-s -w" -o dist/ssh-vault_linux_arm64  ./cmd/ssh-vault && \
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/ssh-vault_darwin_amd64 ./cmd/ssh-vault && \
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/ssh-vault_darwin_arm64 ./cmd/ssh-vault && \
    cd dist && sha256sum ssh-vault_* > checksums.txt

FROM alpine:3

RUN adduser -D -u 1000 vault
WORKDIR /app
COPY --from=builder /build/ssh-vault .
COPY --from=builder /build/dist /dist
RUN mkdir /data && chown vault:vault /data

USER vault
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/ssh-vault", "hub", "--data", "/data", "--dist-dir", "/dist"]
