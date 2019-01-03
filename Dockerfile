FROM golang:latest

WORKDIR /go/src/needu-crawler
COPY . .
RUN go get -v ../...
RUN go install
CMD /go/bin/needu-crawler