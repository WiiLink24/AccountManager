FROM golang:1.25-alpine as builder

# We assume only git is needed for all dependencies.
# openssl is already built-in.
RUN apk add -U --no-cache git

WORKDIR /AccountManager/

# Cache pulled dependencies if not updated.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy all source
COPY *.go .
COPY middleware ./middleware

# Build to name "app".
RUN go build -o app .

FROM alpine:latest

WORKDIR /AccountManager

COPY templates ./templates
COPY assets ./assets
COPY --from=builder /AccountManager/app .

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:9011/health || exit 1

EXPOSE 9011
CMD ["./app"]