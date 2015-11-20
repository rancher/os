FROM golang:1.3.3-cross
MAINTAINER peter.edge@gmail.com

RUN \
	go get -v code.google.com/p/go.tools/cmd/vet && \
	go get -v github.com/golang/lint/golint

RUN mkdir -p /go/src/github.com/fsouza/go-dockerclient
ADD . /go/src/github.com/fsouza/go-dockerclient
WORKDIR /go/src/github.com/fsouza/go-dockerclient
