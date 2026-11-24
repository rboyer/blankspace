FROM alpine:3.13

ARG TARGETOS
ARG TARGETARCH

RUN addgroup blankspace && adduser -S -G blankspace blankspace
COPY blankspace.${TARGETOS}.${TARGETARCH} /bin/blankspace
USER blankspace
ENTRYPOINT ["/bin/blankspace"]
