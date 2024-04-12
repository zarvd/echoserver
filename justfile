default:
    just --list

# Build the project
build:
    cargo build

# Format code with rust
fmt:
    cargo fmt

# Run unit tests against the current platform
unit-test:
    cargo nextest run
    cargo test --doc

# Lint code with clippy
lint:
    cargo fmt --all -- --check
    cargo clippy --all-targets --all-features

# Clean workspace
clean:
    cargo clean

image_repo := "ghcr.io/zarvd/echoserver"
# Build docker image
image tag: clean
    docker buildx build ./ \
        --output=type=docker \
        --no-cache \
        --force-rm \
        --tag {{ image_repo }}:{{ tag }} \
        --file Dockerfile
