use std::net::SocketAddr;
use std::sync::Arc;

use anyhow::Result;
use tokio::net::UdpSocket;
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
    info!("Serving UDP on {}", addr);

    let socket = Arc::new(UdpSocket::bind(addr).await?);

    loop {
        let mut buf = [0; 2048];
        match socket.recv_from(&mut buf).await {
            Ok((n, remote_addr)) => {
                handle(socket.clone(), remote_addr, &buf[0..n]).await?;
            }
            Err(e) => {
                error!("Failed to receive data: {}", e);
                break;
            }
        }
    }
    Ok(())
}
