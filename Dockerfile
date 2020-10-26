FROM golang:1.14-alpine3.12 AS BUILDER
WORKDIR /app

# download build dependencies
RUN apk add nodejs bash make

# build the app
COPY . ./
RUN make build \
    GO_ARGS="CGO_ENABLED=0" \
    LINK="cp -R"

# copy build to a minimal image
FROM scratch
WORKDIR /app
COPY --from=BUILDER /app/build ./
ENTRYPOINT [ "/app/main" ]
