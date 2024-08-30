FROM rust:1.80-bookworm AS builder

WORKDIR /workspace
COPY . .

RUN cargo build --release

FROM debian:bookworm-slim

COPY --from=builder /workspace/target/release/echoserver /usr/local/bin/echoserver

ENTRYPOINT ["/usr/local/bin/echoserver"]
