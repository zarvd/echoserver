FROM rust:1.77-buster AS builder

WORKDIR /workspace
COPY . .

RUN cargo build --release

FROM debian:buster-slim

COPY --from=builder /workspace/target/release/echoserver /usr/local/bin/echoserver

ENTRYPOINT ["/usr/local/bin/echoserver"]
