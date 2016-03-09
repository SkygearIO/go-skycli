FROM golang:1.5

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
RUN go get github.com/tools/godep
COPY Godeps /go/src/app/Godeps
RUN go get github.com/inconshreveable/mousetrap && \
    $GOPATH/bin/godep restore

COPY . /go/src/app

RUN go-wrapper download && \
    go-wrapper install

CMD ["go-wrapper", "run"]

