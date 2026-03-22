FROM golang:1-alpine AS base

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# --- parallel build stages ---

FROM base AS build-hub
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/ssh-vault ./cmd/ssh-vault

FROM base AS build-linux-amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/ssh-vault_linux_amd64 ./cmd/ssh-vault

FROM base AS build-darwin-arm64
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o /out/ssh-vault_darwin_arm64 ./cmd/ssh-vault

# --- assemble dist ---

FROM alpine:3 AS dist
COPY --from=build-linux-amd64 /out/ /dist/
COPY --from=build-darwin-arm64 /out/ /dist/
RUN cd /dist && sha256sum ssh-vault_* > checksums.txt

# --- final image ---

FROM alpine:3

RUN adduser -D -u 1000 vault
WORKDIR /app
COPY --from=build-hub /out/ssh-vault .
COPY --from=dist /dist /dist
RUN mkdir /data && chown vault:vault /data

USER vault
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/ssh-vault", "hub", "--data", "/data", "--dist-dir", "/dist"]
