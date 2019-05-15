FROM golang:1.12

WORKDIR $GOPATH/src/github.com/tadoku/api

COPY . .

RUN GO111MODULE=on go mod download

RUN GO111MODULE=on go install -v ./...

RUN export $(grep -v '^#' .env | xargs)
RUN migrate

ENTRYPOINT ["server"]
