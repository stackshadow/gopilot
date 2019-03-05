FROM alpine:latest

# install openssl that gopilot can create a certificate for you
RUN apk add --no-cache openssl

# We nee a user, the busybox-way
RUN adduser -D -H -u 1000 gorunner 

# Copy our static executable.
COPY gopilot /app/gopilot

# Create an config and www dir
RUN mkdir -p /app/config && \
touch /app/config/core.json && \
mkdir -p /app/www/gopilot

# start
COPY deploy/entrypoint.bash /app/entrypoint.bash
RUN chown -R 1000:0 /app && \
chmod -R u=rwX,g=rX,o= /app && \
chmod a+rx /app/gopilot /app/entrypoint.bash


# Use an unprivileged user.
USER 1000
WORKDIR /app
ENTRYPOINT [ "/bin/sh", "/app/entrypoint.bash"]
CMD ["/app/gopilot"]