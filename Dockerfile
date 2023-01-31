FROM rust:1.67-bullseye AS builder

WORKDIR /workspace
COPY . .

RUN cargo build --release

FROM debian:bullseye-slim

COPY --from=builder /workspace/target/release/echoserver /usr/local/bin/echoserver

ENTRYPOINT ["/usr/local/bin/echoserver"]