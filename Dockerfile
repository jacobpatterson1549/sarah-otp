FROM golang:1.14-alpine3.12 AS BUILDER
RUN apk add nodejs bash make
WORKDIR /app
COPY . ./
RUN make all \
    GO_ARGS="CGO_ENABLED=0" \
    LINK="cp -R"

FROM scratch
WORKDIR /app
COPY --from=BUILDER /app/build ./
ENTRYPOINT [ "/app/main" ]
