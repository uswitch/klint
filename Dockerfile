FROM scratch

ADD bin/klint /klint

ENTRYPOINT ["/klint"]
