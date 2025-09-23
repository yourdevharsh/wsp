use anyhow::{Ok, Result};
use clap::Parser;
use futures_util::SinkExt;
use futures_util::StreamExt;
use live_server::{Options, listen};
use std::io::{BufRead, BufReader, Write};
use std::process::{Command, Stdio};
use std::sync::{Arc, Mutex};
use tokio::net::TcpListener;
use tokio::sync::broadcast;
use tokio_tungstenite::accept_async;
use tokio_tungstenite::tungstenite::Message;

#[derive(Parser)]
struct Cli {
    cmd: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    let args = Cli::parse();

    if args.cmd == "start" {
        println!("Working");

        // Spawn Go node
        let mut child = Command::new("go")
            .args(&["run", "main.go"]) // adjust if compiled
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .expect("failed to start go node");

        let go_stdin = Arc::new(Mutex::new(child.stdin.take().unwrap()));
        let go_stdout = child.stdout.take().unwrap();

        // Channel for broadcasting Go events to all WS clients
        let (tx, _rx) = broadcast::channel::<String>(16);

        // Thread: read Go stdout, forward to WS clients
        let tx_clone = tx.clone();
        std::thread::spawn(move || {
            let reader = BufReader::new(go_stdout);
            for line in reader.lines() {
                if let std::result::Result::Ok(json_line) = line {
                    println!("Go event: {}", json_line);
                    let _ = tx_clone.send(json_line);
                }
            }
        });

        // Task: WebSocket server
        let ws_tx = tx.clone();
        let go_stdin_clone = go_stdin.clone();
        tokio::spawn(async move {
            let listener = TcpListener::bind("127.0.0.1:3001").await.unwrap();
            println!("WebSocket running at ws://127.0.0.1:3001/ws");

            loop {
                let (stream, _) = listener.accept().await.unwrap();
                let peer_tx = ws_tx.clone();
                let mut peer_rx = peer_tx.subscribe();
                let go_stdin_ws = go_stdin_clone.clone();

                tokio::spawn(async move {
                    let ws_stream = accept_async(stream).await.unwrap();
                    let (mut ws_sender, mut ws_receiver) = ws_stream.split();

                    // Task: forward Go events -> WebSocket client
                    let forward_task = tokio::spawn(async move {
                        while let std::result::Result::Ok(msg) = peer_rx.recv().await {
                            if ws_sender.send(Message::Text(msg.into())).await.is_err() {
                                break;
                            }
                        }
                    });

                    // Task: forward WebSocket messages -> Go stdin
                    let recv_task = tokio::spawn(async move {
                        while let Some(std::result::Result::Ok(msg)) = ws_receiver.next().await {
                            if let Message::Text(text) = msg {
                                let mut stdin = go_stdin_ws.lock().unwrap();
                                let _ = stdin.write_all(format!("{}\n", text).as_bytes());
                            }
                        }
                    });

                    tokio::select! {
                        _ = forward_task => (),
                        _ = recv_task => (),
                    }
                });
            }
        });

        // Static file server (UI)
        let server = listen("127.0.0.1:8080", "./public/")
            .await
            .map_err(|e| anyhow::anyhow!("{}", e))?;

        server
            .start(Options::default())
            .await
            .map_err(|e| anyhow::anyhow!("{}", e))?;
    } else {
        println!("Wrong Command");
    }

    Ok(())
}
