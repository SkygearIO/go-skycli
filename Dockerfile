FROM golang:1.5.1

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Copy a minimal set of files to restore Go dependencies to get advantage
# of Docker build cache
RUN go get github.com/tools/godep && \
    go get golang.org/x/tools/cmd/stringer
COPY Godeps /go/src/app/Godeps
RUN $GOPATH/bin/godep restore

COPY . /go/src/app

RUN go-wrapper download && \
    go-wrapper install && \
    go generate ./... && \
    go-wrapper install

CMD ["go-wrapper", "run"]

