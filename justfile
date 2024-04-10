image_repo := "ghcr.io/zarvd/echoserver"

default:
    just --list

build:
    cargo build --release

run:
    cargo run --release

fmt:
    cargo fmt

test:
    cargo test --release

lint:
    cargo clippy --release

clean:
    cargo clean

image tag: clean
    docker buildx build ./ \
        --output=type=docker \
        --no-cache \
        --force-rm \
        --tag {{ image_repo }}:{{ tag }} \
        --file Dockerfile
