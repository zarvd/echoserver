use std::convert::Infallible;
use std::net::SocketAddr;

use anyhow::Result;
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Request, Response, Server};
use tracing::{info, instrument};

#[instrument]
async fn handle(req: Request<Body>) -> Result<Response<Body>, Infallible> {
    Ok(Response::new(req.into_body()))
}

pub async fn serve(addr: SocketAddr) -> Result<()> {
    info!("Serving HTTP on {}", addr);

    let make_svc = make_service_fn(|_conn| async { Ok::<_, Infallible>(service_fn(handle)) });

    Server::bind(&addr).serve(make_svc).await?;

    Ok(())
}
