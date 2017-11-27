FROM golang:alpine
RUN apk --no-cache add ca-certificates git
WORKDIR /go/src/app
COPY . /go/src/app/

RUN go-wrapper download ./...
RUN for binary in confluence-poster git-release-info release-info-confluence; do go build -o bin/${binary} ${binary}/main.go; done; chmod 755 bin/*

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=0 /go/src/app/bin/* /
RUN printf '#!/bin/sh\n\
set -e\n\
echo $@\n\
REPO=$1\n\
shift\n\
PARENTID=$1\n\
shift\n\
/git-release-info $REPO $@ | /release-info-confluence | /confluence-poster $PARENTID\n\
' > /entrypoint.sh; chmod 755 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["--help"]
