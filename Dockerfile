FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY bin/* /bin/
RUN printf '#!/bin/sh\n\
set -e\n\
echo $@\n\
REPO=$1\n\
shift\n\
PARENTID=$1\n\
shift\n\
git-release-info $REPO $@ | release-info-confluence | confluence-poster $PARENTID\n\
' > /entrypoint.sh; chmod 755 /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["--help"]
