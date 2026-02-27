# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /src

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "\
      -s -w \
      -X github.com/mstephenholl/gitops-demo/internal/version.Tag=${VERSION} \
      -X github.com/mstephenholl/gitops-demo/internal/version.Commit=${COMMIT} \
      -X github.com/mstephenholl/gitops-demo/internal/version.BuildTime=${BUILD_TIME}" \
    -o /bin/server ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

COPY --from=builder /bin/server /usr/local/bin/server

USER app

EXPOSE 8080

ENTRYPOINT ["server"]
