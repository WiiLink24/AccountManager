FROM golang:1.23.3-alpine3.19 as builder

# We assume only git is needed for all dependencies.
# openssl is already built-in.
RUN apk add -U --no-cache git

WORKDIR /AccountManager/

# Cache pulled dependencies if not updated.
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy necessary parts of the Mail-Go source into builder's source
COPY *.go ./
COPY middleware middleware
COPY assets assets
COPY templates templates

# Build to name "app".
RUN go build -o app .

EXPOSE 8080
# Wait until there's an actual MySQL connection we can use to start.
CMD ["./app"]