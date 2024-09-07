use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;

use anyhow::Result;
use tokio::net::UdpSocket;
use tokio::time::timeout;
use tracing::{error, info, instrument};

#[instrument(skip_all, fields(
    local_addr=socket.local_addr().unwrap().to_string(),
    remote_addr=remote_addr.to_string(),
))]
async fn handle(socket: Arc<UdpSocket>, remote_addr: SocketAddr, data: &[u8]) -> Result<()> {
    info!("Read {} bytes, writing back", data.len());
    socket.send_to(data, remote_addr).await?;
    Ok(())
}

pub async fn serve(addr: SocketAddr) -> Result<()> {
    info!("Serving UDP on {addr}");

    let socket = timeout(Duration::from_secs(5), UdpSocket::bind(addr))
        .await
        .unwrap_or_else(|e| panic!("timeout to bind UDP server on {addr}: {e}"))
        .map_or_else(
            |e| panic!("failed to bind UDP server on {addr}: {e}"),
            Arc::new,
        );

    info!("UDP server up on {addr}");

    loop {
        let mut buf = [0; 2048];
        match socket.recv_from(&mut buf).await {
            Ok((n, remote_addr)) => {
                handle(Arc::clone(&socket), remote_addr, &buf[0..n])
                    .await
                    .unwrap_or_else(|e| panic!("failed to echo data from UDP server {addr}: {e}"));
            }
            Err(e) => {
                error!("[UDP/{addr}] Failed to receive data: {e}");
            }
        }
    }
}
