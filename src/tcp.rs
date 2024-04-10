use std::net::SocketAddr;

use anyhow::Result;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tracing::{error, info, instrument};

#[instrument(skip_all, fields(
    local_addr=socket.local_addr().unwrap().to_string(),
    remote_addr=socket.peer_addr().unwrap().to_string(),
))]
async fn handle(mut socket: TcpStream) -> Result<()> {
    info!("Accepting new socket");

    let mut buf = vec![0; 1024];

    // In a loop, read data from the socket and write the data back.
    loop {
        let n = socket.read(&mut buf).await?;
        if n == 0 {
            break;
        }

        socket.write_all(&buf[0..n]).await?;
    }

    info!("Dropping socket");

    Ok(())
}

pub async fn serve(addr: SocketAddr) -> Result<()> {
    info!("Serving TCP on {}", addr);

    let lis = TcpListener::bind(addr)
        .await
        .expect(&format!("bind TCP server on {addr}"));

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
