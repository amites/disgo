FROM golang:alpine as builder

RUN apk update && apk add git

RUN go get github.com/dispatchlabs/disgo
RUN cd /go/src/github.com/dispatchlabs/disgo \
  && go build

# -----------------------
FROM alpine:latest

RUN ["mkdir", "-p", "/go-binaries"]

COPY --from=builder /go/src/github.com/dispatchlabs/disgo/disgo /go-binaries/disgo
ADD config /go-binaries/config

WORKDIR /go-binaries

ENTRYPOINT ["/go-binaries/disgo"]
