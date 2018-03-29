FROM golang:1.9
WORKDIR /go/src/github.com/halkeye/git-version-commits
COPY . .
RUN make all

FROM alpine:3.7
COPY --from=0 /go/src/github.com/halkeye/git-version-commits/bin/* /usr/local/bin/

