FROM golang:1.14-buster AS BUILDER

WORKDIR /app

# download node, bash to run wasm tests
RUN apt-get update \
    && apt-get install \
        --no-install-recommends \
        -y \
            nodejs=10.21.0~dfsg-1~deb10u1 \
            wamerican-large=2018.04.16-1

# build the application without static libraries and copy resources instead of linking
COPY . ./
RUN make build \
    GO_ARGS="CGO_ENABLED=0" \
    LINK="cp -R" \
    -j 2

# copy files to a minimal build image
FROM scratch
WORKDIR /app
COPY --from=BUILDER /app/build /app/
CMD [ "/app/main" ]
