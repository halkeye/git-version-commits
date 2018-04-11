FROM golang:1.9
WORKDIR /go/src/github.com/halkeye/git-version-commits
COPY . .
RUN apk add --no-cache ca-certificates
RUN make all

FROM alpine:3.7
RUN apk add --no-cache ca-certificates
COPY --from=0 /go/src/github.com/halkeye/git-version-commits/bin/* /usr/local/bin/

