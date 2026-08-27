# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
RUN apk --no-cache add git
ARG VERSION=dev
COPY hubtojo/go.mod hubtojo/go.sum ./
RUN go mod download
RUN --mount=target=. \
  mkdir -p /build && \
  cd hubtojo && \
  CGO_ENABLED=0 \
  go build -ldflags "-X main.Version=${VERSION}" -a -installsuffix cgo \
  -o /build/hubtojo \
  .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the pre-built binary from the previous stage
COPY --from=builder /build/hubtojo .

# COPY the entrypoint script that will load certificates and run the binary
COPY docker/entrypoint.sh .

ENTRYPOINT ["/app/entrypoint.sh"]

EXPOSE 8080

# Run the executable
CMD ["./hubtojo"]
