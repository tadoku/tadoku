# Build server
FROM golang:1.14.2 as build
WORKDIR /base
COPY . .
RUN GO111MODULE=on go mod download
RUN GO111MODULE=on go install -v ./...

# Create production container
FROM alpine:3.7
COPY --from=build /go/bin/server /go/bin/migrate /usr/bin/

# Run the app
ENTRYPOINT ["server"]
