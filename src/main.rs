mod http;
mod signal;
mod tcp;
mod udp;

use std::collections::HashSet;
use std::fmt::{Display, Formatter};
use std::net::{IpAddr, Ipv4Addr, SocketAddr};

use anyhow::Result;
use clap::Parser;
use tokio::runtime;
use tracing::{info, Level};
use tracing_subscriber::fmt::format::FmtSpan;

#[derive(Default, Debug, Clone)]
struct SocketAddrs(Vec<SocketAddr>);

impl Display for SocketAddrs {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        let mut ports = [false; u16::MAX as usize];
        for addr in &self.0 {
            ports[addr.port() as usize] = true;
        }
        let mut ranges = vec![];
        let mut from: Option<u16> = None;
        for (port, enabled) in ports.into_iter().enumerate() {
            let port = port as u16;
            if !enabled {
                if let Some(from_port) = from {
                    ranges.push((from_port, port - 1));
                    from = None;
                }
            } else if from.is_none() {
                from = Some(port);
            }
        }

        let ranges: Vec<_> = ranges
            .into_iter()
            .map(|(from, end)| {
                if from == end {
                    from.to_string()
                } else {
                    format!("{from}-{end}")
                }
            })
            .collect();

        write!(f, "{}", ranges.join(","))
    }
}

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct App {
    #[arg(long = "tcp-ports", value_parser = parse_socket_addrs, default_value_t = SocketAddrs::default())]
    tcp_addrs: SocketAddrs,

    #[arg(long = "udp-ports", value_parser = parse_socket_addrs, default_value_t = SocketAddrs::default())]
    udp_addrs: SocketAddrs,

    #[arg(long = "http-ports", value_parser = parse_socket_addrs, default_value_t = SocketAddrs::default())]
    http_addrs: SocketAddrs,
}

fn parse_socket_addrs(arg: &str) -> Result<SocketAddrs> {
    if arg.is_empty() {
        return Ok(SocketAddrs::default());
    }
    let ranges: Vec<_> = arg
        .split(',')
        .map(|s| s.split_once('-').unwrap_or((s, s)))
        .map(|(from, end)| (from.parse::<u16>(), end.parse::<u16>()))
        .collect();
    let mut ports = HashSet::new();
    for (from, end) in ranges {
        for p in from?..=end? {
            ports.insert(p);
        }
    }
    let addrs = ports
        .iter()
        .map(|&port| SocketAddr::new(IpAddr::V4(Ipv4Addr::new(0, 0, 0, 0)), port))
        .collect();
    Ok(SocketAddrs(addrs))
}

fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_max_level(Level::INFO)
        .with_span_events(FmtSpan::CLOSE)
        .init();
    let app: App = App::parse();

    let rt = runtime::Builder::new_multi_thread().enable_io().build()?;
    rt.block_on(async {
        let mut handles = vec![];
        for addr in app.tcp_addrs.0 {
            handles.push(tokio::spawn(tcp::serve(addr)));
        }
        for addr in app.udp_addrs.0 {
            handles.push(tokio::spawn(udp::serve(addr)));
        }
        for addr in app.http_addrs.0 {
            handles.push(tokio::spawn(http::serve(addr)));
        }

        if !handles.is_empty() {
            signal::shutdown().await;
        }
        info!("Shutting down");
    });

    Ok(())
}
