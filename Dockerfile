FROM golang:1.14-alpine3.12 AS BUILDER
WORKDIR /app

# download node, bash to run wasm tests, make to build
RUN apk add nodejs bash make

# build the application without static libraries and copy resources instead of linking
COPY . ./
RUN make build \
    GO_ARGS="CGO_ENABLED=0" \
    LINK="cp -R" \
    -j 2

# copy files to a minimal build image
FROM scratch
WORKDIR /app
COPY --from=BUILDER /app/build ./
CMD [ "/app/main", \
    "-tls-cert-file=/app/tls-cert.pem", \
    "-tls-key-file=/app/tls-key.pem" ]
