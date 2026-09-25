FROM gcr.io/distroless/static-debian13:nonroot

COPY ./dist/mtls /mtls

USER 1000:1000

ENTRYPOINT ["/mtls"]
