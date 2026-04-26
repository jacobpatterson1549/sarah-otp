FROM golang:1.26-alpine3.23 AS BUILDER
WORKDIR /app
RUN apk add --no-cache \
        make=~4.4.1-r3 \
        bash=~5.3.3-r1 \
        nodejs=~24.14.1-r0
COPY . ./
RUN make all \
    GO_ARGS="CGO_ENABLED=0" \
    LINK="cp -R"

FROM scratch
WORKDIR /app
COPY --from=BUILDER /app/build ./
ENTRYPOINT [ "/app/main" ]
