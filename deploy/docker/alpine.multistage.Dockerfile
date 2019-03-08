############################
# Builder
############################
#FROM golang:alpine AS builder
FROM golang:1.10-alpine AS builder

# install ca-certificates
# binutils = strip
RUN apk update && \
    apk add --no-cache git binutils ca-certificates && update-ca-certificates

# Create gorunner
# the default way
# RUN useradd -M gorunner && \
# usermod -L gorunner &&

# The busybox-way
RUN adduser -D -H -u 1000 gorunner 

COPY ./src $GOPATH/src/
WORKDIR $GOPATH

# Using go get.
RUN go get -d -v ./src

# Build the binary.
RUN mkdir -p /app && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/gopilot ./src && \
    echo "Strip gopilot $(stat -c %s /app/gopilot) bytes" && \
    strip /app/gopilot && \
    echo "Striped gopilot $(stat -c %s /app/gopilot) bytes"

############################
# Runtime image
############################
FROM alpine:latest

# install openssl that gopilot can create a certificate for you
RUN apk add --no-cache openssl

# we need certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# We created a user inside the builder, copy it
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable.
COPY --from=builder /app/gopilot /app/gopilot

# Create an config and www dir
RUN mkdir -p /app/config && \
    touch /app/config/core.json && \
    mkdir -p /app/www/gopilot

# start
COPY deploy/docker/entrypoint.bash /entrypoint
RUN chown -R 1000:0 /app && \
    chmod -R u=rwX,g=rX,o= /app && \
    chmod a+rx /app/gopilot /entrypoint


# Use an unprivileged user.
USER 1000
WORKDIR /app
ENTRYPOINT [ "/bin/sh", "/entrypoint"]
CMD [ "-websocket", "-nftSkipOnStart", "-websocket.addr", ":3333"]