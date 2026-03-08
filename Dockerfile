FROM  gcr.io/distroless/static:nonroot

COPY relay /relay

CMD ["/relay"]