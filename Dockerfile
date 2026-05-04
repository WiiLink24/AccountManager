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

EXPOSE 9011
CMD ["./app"]