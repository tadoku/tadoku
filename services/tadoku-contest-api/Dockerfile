# Build server
FROM golang:1.16-alpine as build
WORKDIR /base
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go install -v ./...

# Create production container
FROM alpine:3.7
COPY --from=build /go/bin/server /go/bin/migrate /
COPY --from=build /base/migrations /migrations/

# Run the app
ENTRYPOINT ["/server"]
