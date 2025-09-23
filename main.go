package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	ma "github.com/multiformats/go-multiaddr"
)

// Event struct
type Event struct {
	Type    string `json:"type"`
	From    string `json:"from,omitempty"`
	Content string `json:"content,omitempty"`
	PeerID  string `json:"peerId,omitempty"`
	Error   string `json:"error,omitempty"`
}

func emit(event Event) {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}

func main() {
	ctx := context.Background()

	// --- Create the libp2p node with DHT and Relay capabilities ---
	var kademliaDHT *dht.IpfsDHT
	node, err := libp2p.New(
		// Use the DHT as the routing mechanism
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			kademliaDHT, err = dht.New(ctx, h)
			return kademliaDHT, err
		}),
		// Run the DHT in server mode to help the network
		libp2p.EnableRelay(),
		// Enable hole punching for NAT traversal
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		panic(err)
	}
	emit(Event{Type: "started", PeerID: node.ID().String()})

	// --- Connect to Bootstrap Peers ---
	emit(Event{Type: "status", Content: "Connecting to bootstrap peers..."})
	var wg sync.WaitGroup
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := node.Connect(ctx, *pi); err != nil {
				emit(Event{Type: "error", Error: fmt.Sprintf("Bootstrap connect error: %s", err)})
			} else {
				emit(Event{Type: "connected", PeerID: pi.ID.String()})
			}
		}()
	}
	wg.Wait()

	// GossipSub Setup
	ps, err := pubsub.NewGossipSub(ctx, node)
	if err != nil {
		panic(err)
	}

	// --- Automated peer discovery ---
	emit(Event{Type: "status", Content: "Announcing ourselves..."})
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)

	dutil.Advertise(ctx, routingDiscovery, "chat-room")
	emit(Event{Type: "status", Content: "Searching for other peers..."})

	// Find peers and connect to them
	go func() {
		peerChan, err := routingDiscovery.FindPeers(ctx, "chat-room")
		if err != nil {
			panic(err)
		}
		for peer := range peerChan {
			// Don't connect to ourselves
			if peer.ID == node.ID() {
				continue
			}
			emit(Event{Type: "status", Content: fmt.Sprintf("Found peer: %s. Connecting...", peer.ID.String())})
			if err := node.Connect(ctx, peer); err != nil {
				emit(Event{Type: "error", Error: fmt.Sprintf("Connect to peer %s failed: %s", peer.ID.String(), err)})
			} else {
				emit(Event{Type: "connected", PeerID: peer.ID.String()})
			}
		}
	}()

	// Join the chat room topic
	topic, _ := ps.Join("chat-room")
	sub, _ := topic.Subscribe()

	// Goroutine for listening to messages
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				emit(Event{Type: "error", Error: err.Error()})
				continue
			}
			if msg.ReceivedFrom == node.ID() {
				continue // skip own messages
			}
			emit(Event{
				Type:    "message",
				From:    msg.ReceivedFrom.String(),
				Content: string(msg.Data),
			})
		}
	}()

	// Read commands from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {

		case "connect":
			if len(parts) < 2 {
				emit(Event{Type: "error", Error: "Usage: connect <multiaddr>"})
				continue
			}
			maddr, err := ma.NewMultiaddr(parts[1])
			if err != nil {
				emit(Event{Type: "error", Error: "Invalid multiaddr"})
				continue
			}
			pi, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				emit(Event{Type: "error", Error: "Invalid peer info"})
				continue
			}
			if err := node.Connect(ctx, *pi); err != nil {
				emit(Event{Type: "error", Error: err.Error()})
			} else {
				emit(Event{Type: "connected", PeerID: pi.ID.String()})
			}
		case "send":
			if len(parts) < 2 {
				emit(Event{Type: "error", Error: "Usage: send <message>"})
				continue
			}
			msg := parts[1]
			if err := topic.Publish(ctx, []byte(msg)); err != nil {
				emit(Event{Type: "error", Error: "Failed to publish"})
			} else {
				emit(Event{Type: "sent", Content: msg})
			}
		default:
			emit(Event{Type: "error", Error: "Unknown command"})
		}
	}
}
