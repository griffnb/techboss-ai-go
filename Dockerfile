#This is the docker file for final cloudformation image
FROM alpine:latest as build
RUN apk --update add ca-certificates && apk add tzdata


# Start from a small base
FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo


# Our application requires no privileges
# so run it with a non-root user
#ADD users /etc/passwd
#USER nobody

# Our application runs on port 8080
# so allow hosts to bind to that port
EXPOSE 8080

# Add our application binary
ADD app /app

# Run our application!
ENTRYPOINT [ "/app" ]
