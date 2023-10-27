FROM alpine:3.13
RUN addgroup blankspace && adduser -S -G blankspace blankspace
COPY blankspace /bin/blankspace
USER blankspace
ENTRYPOINT ["/bin/blankspace"]
