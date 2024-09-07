use std::net::SocketAddr;
use std::time::{Duration, Instant};

use anyhow::Result;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tokio::time::timeout;
use tracing::{error, info, instrument};

#[instrument(skip_all, fields(
    local_addr=socket.local_addr().unwrap().to_string(),
    remote_addr=socket.peer_addr().unwrap().to_string(),
))]
async fn handle(mut socket: TcpStream) -> Result<()> {
    info!("Accepting new socket");
    let t1 = Instant::now();

    let mut buf = vec![0; 1024];

    // In a loop, read data from the socket and write the data back.
    loop {
        let n = socket.read(&mut buf).await?;
        if n == 0 {
            break;
        }

        socket.write_all(&buf[0..n]).await?;
    }

    let elapsed = t1.elapsed();
    info!("Dropping socket after {} ms", elapsed.as_millis());

    Ok(())
}

pub async fn serve(addr: SocketAddr) -> Result<()> {
    info!("Serving TCP on {addr}");

    let lis = timeout(Duration::from_secs(5), TcpListener::bind(addr))
        .await
        .unwrap_or_else(|e| panic!("timeout to bind TCP server on {addr}: {e}"))
        .unwrap_or_else(|e| panic!("failed to bind TCP server on {addr}: {e}"));

    info!("TCP server up on {addr}");

    loop {
        match lis.accept().await {
            Ok((socket, _remote_addr)) => {
                tokio::spawn(handle(socket));
            }
            Err(e) => {
                error!("[TCP/{addr}] Failed to accept new socket: {e}");
            }
        }
    }
}
