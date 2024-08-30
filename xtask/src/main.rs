use std::fmt::Debug;

use anyhow::Result;
use clap::{Parser, Subcommand};
use console::style;
use duct::cmd;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[command(subcommand)]
    action: Action,
}

#[derive(Debug, Subcommand)]
enum Action {
    /// Run linting tools on the codebase
    Lint,
    /// Install required development tools
    InstallTools,
    /// Run benchmarks
    Bench,
    /// Run tests
    Test,
    /// Generate documentation
    Doc,
    /// Format the code
    Fmt,
    /// Build docker image
    BuildDockerImage {
        #[arg(long, default_value = "ghcr.io/zarvd/echoserver")]
        image: String,
        #[arg(long, default_value = "latest")]
        tag: String,
    },
}

fn fmt() -> Result<()> {
    println!("{}", style("cargo +nightly fmt").bold());
    cmd!("cargo", "+nightly", "fmt").run()?;
    Ok(())
}

fn check_fmt() -> Result<()> {
    println!("{}", style("cargo +nightly fmt --check").bold());
    cmd!("cargo", "+nightly", "fmt", "--check").run()?;
    Ok(())
}

fn clippy() -> Result<()> {
    println!(
        "{}",
        style("cargo clippy --all-targets --all-features").bold()
    );
    cmd!("cargo", "clippy", "--all-targets", "--all-features").run()?;
    Ok(())
}

fn unit_test() -> Result<()> {
    println!("{}", style("cargo nextest run --all-features").bold());
    cmd!("cargo", "nextest", "run", "--all-features").run()?;
    Ok(())
}

fn bench() -> Result<()> {
    println!("{}", style("cargo bench --all-features").bold());
    cmd!("cargo", "bench", "--all-features").run()?;
    Ok(())
}

fn doc() -> Result<()> {
    println!("{}", style("cargo doc --no-deps --all-features").bold());
    cmd!("cargo", "doc", "--no-deps", "--all-features").run()?;
    Ok(())
}

fn build_docker_image(image: String, tag: String) -> Result<()> {
    cmd!(
        "docker",
        "buildx",
        "build",
        "--output=type=docker",
        "--no-cache",
        "--force-rm",
        "--file=Dockerfile",
        format!("--tag={image}:{tag}"),
        "."
    )
    .run()?;
    Ok(())
}

fn main() -> Result<()> {
    let args = Args::parse();

    match args.action {
        Action::InstallTools => {
            println!("{}", style("cargo install cargo-nextest").bold());
            cmd!("cargo", "install", "cargo-nextest", "--locked").run()?;
        }
        Action::Lint => {
            check_fmt()?;
            clippy()?;
        }
        Action::Test => {
            unit_test()?;
        }
        Action::Bench => {
            bench()?;
        }
        Action::Doc => {
            doc()?;
        }
        Action::Fmt => {
            fmt()?;
        }
        Action::BuildDockerImage { image, tag } => {
            build_docker_image(image, tag)?;
        }
    }

    Ok(())
}
