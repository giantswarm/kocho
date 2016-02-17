FROM busybox:ubuntu-14.04

MAINTAINER Stephan Zeissler <stephan@giantswarm.io>

COPY kocho /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/kocho"]