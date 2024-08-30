use std::net::SocketAddr;
use std::time::Duration;

use anyhow::Result;
use axum::extract::Request;
use axum::response::{IntoResponse, Response};
use axum::routing::{get, post};
use axum::Router;
use tokio::net::TcpListener;
use tower_http::cors::CorsLayer;
use tower_http::timeout::TimeoutLayer;
use tower_http::trace::{DefaultMakeSpan, DefaultOnRequest, DefaultOnResponse, TraceLayer};
use tracing::{info, Level};

async fn ping(_req: Request) -> Response {
    "pong".into_response()
}

async fn echo(req: Request) -> Response {
    let (req_parts, body) = req.into_parts();

    let (mut resp_parts, body) = Response::new(body).into_parts();
    resp_parts.headers = req_parts.headers;

    Response::from_parts(resp_parts, body)
}

pub async fn serve(addr: SocketAddr) -> Result<()> {
    info!("Serving HTTP on {}", addr);

    let app = Router::new()
        .route("/ping", get(ping))
        .route("/echo", post(echo))
        .layer(CorsLayer::permissive())
        .layer(TimeoutLayer::new(Duration::from_secs(10)))
        .layer(
            TraceLayer::new_for_http()
                .make_span_with(
                    DefaultMakeSpan::new()
                        .level(Level::INFO)
                        .include_headers(true),
                )
                .on_request(DefaultOnRequest::new().level(Level::INFO))
                .on_response(DefaultOnResponse::new().level(Level::INFO)),
        );

    let listener = TcpListener::bind(addr)
        .await
        .unwrap_or_else(|e| panic!("failed to bind HTTP server on {addr}: {e}"));
    axum::serve(listener, app)
        .await
        .unwrap_or_else(|e| panic!("failed to serve HTTP server on {addr}: {e}"));

    Ok(())
}
