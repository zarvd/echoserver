use std::net::SocketAddr;

use anyhow::Result;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};
use tracing::{error, info, instrument};

#[instrument(skip_all, fields(
    local_addr=socket.local_addr().unwrap().to_string(),
    remote_addr=socket.peer_addr().unwrap().to_string(),
))]
async fn handle(mut socket: TcpStream) -> Result<()> {
    info!("Accepting new socket");

    let (rh, mut wh) = socket.split();
    let mut rh = BufReader::new(rh);
    loop {
        let mut buf = String::new();
        match rh.read_line(&mut buf).await {
            Ok(n) if n == 0 => break,
            Ok(n) => {
                info!("Read {} bytes, writing back", n);
                wh.write_all(buf.as_bytes()).await?;
            }
            Err(e) => {
                error!("Failed to read from socket: {}", e);
                break;
            }
        }
    }
    info!("Dropping socket");

    Ok(())
}

pub async fn serve(addr: SocketAddr) -> Result<()> {
    info!("Serving TCP on {}", addr);

    let lis = TcpListener::bind(addr).await?;

    loop {
        match lis.accept().await {
            Ok((socket, _remote_addr)) => {
                tokio::spawn(handle(socket));
            }
            Err(e) => {
                error!("[TCP/{}] Failed to accept new socket: {}", addr, e);
            }
        }
    }
}
