FROM golang:alpine AS build-env
RUN apk --no-cache add ca-certificates
COPY go.* main.go /src/
RUN cd /src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main

# final stage
FROM scratch

COPY --from=build-env /src/main /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/main"]
