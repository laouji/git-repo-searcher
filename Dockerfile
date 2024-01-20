FROM golang:1.21
LABEL maintainer="laouji"

RUN go install github.com/cespare/reflex@latest

WORKDIR $GOPATH/src/github.com/laouji/git-repo-searcher

EXPOSE 5000

CMD $GOPATH/bin/sclng-backend-test-v1
