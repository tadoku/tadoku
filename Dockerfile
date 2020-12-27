FROM golang:1.14.2

WORKDIR $GOPATH/src/github.com/tadoku/api

COPY . .

RUN GO111MODULE=on go mod download

RUN GO111MODULE=on go install -v ./...

ENTRYPOINT ["server"]
