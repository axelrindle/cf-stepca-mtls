FROM gcr.io/distroless/static-debian13:nonroot

COPY --chmod=755 ./dist/mtls /mtls

USER 1000:1000

ENTRYPOINT ["/mtls"]
