FROM golang:1.18-bullseye AS builder

WORKDIR /workspace

COPY . .

RUN make build

FROM gcr.io/distroless/static

COPY --from=builder /workspace/bin/echoserver /usr/local/bin/echoserver

ENTRYPOINT [ "/usr/local/bin/echoserver" ]