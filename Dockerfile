FROM golang:alpine AS build-env
RUN apk --no-cache add ca-certificates
COPY . /src/
RUN cd /src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o coinbase .

# # final stage
FROM scratch

COPY --from=build-env /src/coinbase /
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/coinbase", "serve"]
