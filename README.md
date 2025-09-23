# wsp 🌐

A peer-to-peer chat application built with Go, Rust, and WebSockets. This project uses a `go-libp2p` node for the decentralized backend and a Rust application to bridge the P2P network to a simple web-based UI via WebSockets.

---

## 🚀 About The Project

`wsp` demonstrates how to combine the strengths of different technologies to build a decentralized application. The core P2P logic is handled by Go's robust `go-libp2p` implementation, while Rust provides a high-performance wrapper that serves a web UI and communicates with it in real-time using WebSockets.

This project is a great example of:
* Inter-process communication between Rust and Go.
* Bridging a P2P network (libp2p) to the conventional web (WebSockets).
* Creating a simple, decentralized chat application from scratch.

### Architecture

The application's architecture is composed of three main parts:

1.  **Go P2P Node (`main.go`):** The heart of the application. It connects to the libp2p network, discovers peers using a Kademlia DHT, and uses a Pub/Sub topic to send and receive chat messages. It communicates with the Rust wrapper via `stdin` and `stdout`, sending and receiving simple commands and JSON events.

2.  **Rust Bridge (`main.rs`):** This acts as the orchestrator. It spawns the Go program as a child process, manages its I/O, and hosts a WebSocket server. It forwards messages between the web UI and the Go node, effectively acting as a bridge between the two. It also serves the static `index.html` file.

3.  **Frontend (`index.html`):** A minimal HTML file with JavaScript that connects to the Rust WebSocket server to provide a real-time user interface for the chat.


---

### Prerequisites

To run this project, you will need to have the following installed on your system:

* [Go](https://go.dev/doc/install)
* [Rust](https://www.rust-lang.org/tools/install) (with Cargo)

---

## ⚙️ Getting Started

Follow these steps to get the application up and running.

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/your-username/wsp.git](https://github.com/your-username/wsp.git)
    cd wsp
    ```

2.  **Run the application:**
    Use Cargo to run the Rust application. The Rust code is configured to automatically build and execute the Go program.
    ```bash
    cargo run -- start
    ```

3.  **Open the UI:**
    Once the server is running, you will see output in your console indicating that the WebSocket and web servers have started. Open your web browser and navigate to:
    ```
    [http://127.0.0.1:8080](http://127.0.0.1:8080)
    ```

You should now see the chat interface. Open multiple browser tabs to simulate different users. The P2P node will automatically discover other peers on the network and relay messages between them.

---

## 🔧 How It Works

* The Rust application is started with `cargo run -- start`.
* It spawns the Go P2P node (`go run main.go`) as a child process.
* The Rust app listens to the `stdout` of the Go process to receive JSON events (e.g., new messages, peer connections).
* These events are broadcasted to all connected WebSocket clients (browser UIs).
* When a user sends a message from the web UI, the message is sent over the WebSocket to the Rust server.
* The Rust server writes the message as a command to the `stdin` of the Go process.
* The Go process receives the command, publishes the message to the libp2p Pub/Sub topic, and the cycle continues.

---

## 🤝 Contributing

Contributions are welcome! If you have ideas for improvements, please open an issue or submit a pull request.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

---
