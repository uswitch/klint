FROM alpine:3 as base

RUN apk add -U --no-cache ca-certificates

FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD --chmod=755  bin/klint /klint

ENTRYPOINT ["/klint"]