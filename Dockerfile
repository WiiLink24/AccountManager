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
COPY . .

# Build to name "app".
RUN go build -o app .

FROM alpine:latest

WORKDIR /AccountManager

COPY --from=builder /AccountManager/app .
COPY templates ./templates
COPY assets ./assets

CMD ["./app"]