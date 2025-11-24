# 🌐 wsp

A proof-of-concept project demonstrating a three-tier architecture for a decentralized, real-time chat application using `libp2p` (Go) for networking, orchestrated by an asynchronous Rust backend, and exposed via a simple HTML/JavaScript frontend.

## 🌟 Intro

This project serves as a comprehensive example of how to decouple the core decentralized networking logic from the user interface and orchestration layer. The result is a robust P2P chat client that leverages the battle-tested **GossipSub** protocol via `go-libp2p`, wrapped in a user-friendly WebSocket gateway written in Rust.

## 🛠️ How It Works

The system operates using a tri-layered architecture:

1.  **P2P Network Layer (Go `main.go`):** The core Go binary initializes a `libp2p` node, connects to bootstrap peers, and uses a Distributed Hash Table (DHT) for peer discovery. It joins the global `"chat-room"` topic using **GossipSub** for public messaging. This layer communicates solely through structured JSON objects (events) via standard output (`stdout`) and accepts commands (like `send` and `connect`) via standard input (`stdin`).

2.  **Gateway/Orchestration Layer (Rust `main.rs`):** The Rust executable acts as the system broker. It performs three key functions concurrently:
    * Spawns and manages the Go P2P process.
    * Runs a **WebSocket server** on `127.0.0.1:3001`.
    * **Bridges Communication:** It reads JSON events from the Go process's `stdout` and broadcasts them to all connected WebSocket clients. Conversely, it reads commands from the WebSocket clients and forwards them to the Go process's `stdin`.
    * Serves the static HTML frontend.

3.  **Presentation Layer (HTML/JavaScript `index.html`):** A simple, standard web page that connects to the Rust WebSocket gateway. It sends user messages as commands and updates the UI in real-time based on the JSON events received from the decentralized network.

## ✨ Features

* **Peer-to-Peer Messaging:** Utilizes `libp2p`'s GossipSub protocol for efficient, resilient, and un-censorable chat broadcast.
* **Decentralized Discovery:** Implements Kademlia DHT for automated peer finding within the network using the `"chat-room"` service tag.
* **NAT Traversal:** Automatically enables **Relay** and **Hole Punching** to allow nodes behind NATs or firewalls to connect.
* **Real-time Web UI:** Simple HTML/JS frontend provides a quick way to interact with the P2P network via a WebSocket bridge.
* **Multi-Language Stack:** Combines the networking power of **Go** with the concurrent efficiency of **Rust** and the universal access of **JavaScript**.

## 💻 Tech Stack

| Component | Technology | Role |
| :--- | :--- | :--- |
| **P2P Core** | **Go** (`golang`) | `libp2p` node, GossipSub, Kademlia DHT. |
| **Orchestration** | **Rust** (`tokio`) | Asynchronous WebSocket Server, Inter-Process Communication (IPC), Static File Serving. |
| **Frontend** | **HTML / JavaScript** | Simple UI for sending and receiving messages. |
| **Networking** | **`go-libp2p`** | Core P2P library. |
| **WS Server** | **`tokio-tungstenite`** | WebSocket implementation in Rust. |

## 🚀 Getting Started

To run this project, you will need both the Go and Rust toolchains installed.

### Prerequisites

* **Go:** Version 1.18+
* **Rust:** Stable channel with Cargo

### Setup and Installation

1.  **Clone the repository (or ensure the files are in the same directory):**

    ```bash
    git clone <your-repo-url>
    cd <repo-name>
    ```

2.  **Run the Rust Gateway:**
    The `main.rs` file includes the logic to compile and execute the Go node as a child process, start the WebSocket server, and serve the UI.

    ```bash
    cargo run -- start
    ```

    You should see output similar to:

    ```
    Working
    WebSocket running at ws://127.0.0.1:3001/ws
    Go event: {"type":"started","peerId":"12D3Koo..."}
    ...
    ```

3.  **Usage**
    * Open your web browser and navigate to the UI address, typically **`http://127.0.0.1:8080`**.
    * Since this is a P2P application, you must run at least two instances on different machines (or use tools like ngrok/tailscale to expose your ports) to see messages being routed across the network.
    * Type a message into the input field and click "Send". This command is routed: `HTML -> Rust WS -> Go STDIN -> libp2p Network -> other peers`.

## 📐 System Design and Architecture

### Architecture Layers

The design uses a clean separation of concerns, treating the decentralized network as an external service to the application front-end:

1.  **Web Client (Ephemeral):** Handles state and user interaction. Simple and disposable.
2.  **API/Gateway (Rust):** A robust, high-performance, asynchronous IO layer responsible for translation and routing between the synchronous application layer and the web client.
3.  **Application Core (Go):** The single, isolated process dedicated to maintaining the P2P connection state and logic.

### Inter-Process Communication (IPC)

The current IPC mechanism between Rust and Go relies on **Standard Input/Output (STDIN/STDOUT)**, where structured data is passed as JSON.

* **Go (Emitters):** Writes JSON events (`{"type": "message", "content": "..."}`) to `stdout`.
* **Rust (Listeners):** Uses `BufReader` on the Go process's `stdout` to read and parse these JSON lines, which are then broadcast via a `tokio::sync::broadcast` channel to the active WebSocket clients.
* **Rust (Senders):** Locks a mutex-protected handle to the Go process's `stdin` to write command strings (`send <message>`).

### Potential System Improvements

* **Replace STDIN/STDOUT IPC:** Using a more structured protocol like **gRPC** or a simple **HTTP API** between the Rust gateway and the Go core would provide better error handling, clearer schema definitions, and eliminate the overhead of line-by-line JSON parsing from standard streams.
* **Bundle Go Binary:** Instead of using `go run main.go`, the Rust program should execute a pre-compiled Go binary for faster startup and easier deployment.

## 📊 DSA Analysis and Potential Improvements

### Kademlia Distributed Hash Table (DHT)

* **Data Structure:** The DHT implements a key-value store used primarily for peer routing and service discovery. It relies on a special **XOR metric** for distance calculation between peer IDs.
* **Algorithm:** Kademlia provides lookups (for finding peers associated with the `"chat-room"` topic) with a theoretical complexity of **$O(\log N)$**, where $N$ is the number of nodes in the network. This ensures efficient peer discovery even in very large networks.
* **Potential Improvement:** Implement custom record storage on the DHT if the application were to require persistent, decentralized data (e.g., user profiles or stored message history).

### GossipSub

* **Data Structure/Overlay:** GossipSub uses a probabilistic, sparse overlay network built on top of the full P2P mesh. Each peer maintains a small, constant number of connections for a given topic.
* **Algorithm:** Messages are quickly propagated through a "gossip" phase to randomly selected peers and then reliably delivered to a smaller "mesh" network of highly trusted peers. This achieves a balance between the speed and resilience of full flooding and the efficiency of a single-source broadcast.
* **Potential Improvement:** Fine-tuning the GossipSub parameters (D, D\_low, D\_high) to optimize for the specific use case (e.g., prioritize lower latency over bandwidth, or vice-versa) based on network conditions and expected usage patterns.

## 📈 Performance Metrics

In the context of a P2P application, performance is measured by network metrics rather than simple server throughput:

| Metric | Description | Expected Behavior |
| :--- | :--- | :--- |
| **Message Latency** | The time taken for a message to propagate from one peer to all other peers in the chat room. | Low, due to the high connectivity and fan-out of GossipSub. Dependent on the number of hops. |
| **Topic Accuracy/Reach** | The percentage of peers in the topic that eventually receive the message. | Very High (approaching 100%) due to GossipSub's resilience and redundancy mechanism. |
| **Bandwidth Consumption** | Network usage, primarily driven by message propagation and the ongoing DHT/Relay traffic. | Moderate to High, as GossipSub intentionally introduces redundancy to ensure message delivery. |
| **Connection Stability** | Reliability of long-lived peer connections, especially across NATs. | High, due to explicit enabling of `libp2p.EnableRelay()` and `libp2p.EnableHolePunching()`. |

## ⚖️ Trade-offs: Why Use This Architecture?

| Trade-off | Rationale for Current Choice |
| :--- | :--- |
| **Complexity vs. Decentralization** | **Chosen:** High Complexity. Using `libp2p` dramatically increases implementation complexity, but it is the **only way** to achieve true censorship-resistance and resilience, eliminating the need for a central server. |
| **Multi-Stack vs. Single-Stack** | **Chosen:** Multi-Stack (Go/Rust/JS). Go is optimized for robust P2P networking primitives. Rust is chosen for its superior concurrency and safety features, making it an ideal asynchronous gateway for the web. This maximizes performance in each layer. |
| **IPC (STDIO) vs. (gRPC)** | **Chosen:** STDIN/STDOUT. This method is the fastest to prototype and implement, as it requires no serialization/deserialization framework setup, but sacrifices robustness and type safety. |
| **GossipSub vs. Central Broker** | **Chosen:** GossipSub. While a central broker is trivially simple, GossipSub ensures that even if several nodes fail, the chat network remains operational, prioritizing resilience over setup simplicity. |

## 🚀 Impact: Real Use Cases

This project demonstrates the core functionality needed for several next-generation applications:

1.  **Censorship-Resistant Messaging:** Provides a foundation for communication tools that cannot be shut down or censored by a single entity.
2.  **Distributed Gaming:** The P2P PubSub layer can be adapted to broadcast game state updates (e.g., player movements) to all players without relying on a central match-making server.
3.  **Decentralized Data Sharing (IPFS/Filecoin):** The `libp2p` node is already integrated with the IPFS networking layer, allowing for the future integration of decentralized file storage and retrieval alongside the chat functionality.

## 🔮 Future Updates

* **Direct Messaging (DM):** Implement secure, encrypted peer-to-peer streams on a one-to-one basis using the `libp2p` secure channel.
* **Persistence:** Integrate a decentralized storage solution (like a local database or IPFS content addressing) to store messages history, rather than just relaying live messages.
* **Structured IPC (Rust/Go):** Refactor the STDIN/STDOUT communication to use a more robust IPC method like gRPC for better schema validation and reduced runtime errors.
* **WebAssembly Client:** Compile the Go `libp2p` core directly to WebAssembly to run the P2P node natively in the browser, eliminating the need for the Rust WebSocket gateway.
