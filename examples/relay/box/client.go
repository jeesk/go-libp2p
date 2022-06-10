package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
	"log"
)

func main() {

	h1, err := libp2p.New(libp2p.EnableRelay())
	if err != nil {
		log.Printf("Failed to create h1: %v", err)
		return
	}
	releyAddr, err := peer.AddrInfoFromString("/ip4/192.168.1.23/tcp/7676/ipfs/QmZLwcrA4NLWUehrrh6fjC8xunUd7uv1yUbDxkyNYmdzKZ")
	h3, err3 := peer.AddrInfoFromString("/ip6/240e:36a:1490:f100::8a0/tcp/32863/ipfs/QmTq7Ej4gJQT4kchfo1EBu9tMpKEhZNEQQRwJP9A2AZuuU")

	// h1 和h3 连接到h2
	// Connect both h1 and h3 to h2, but not to each other
	if err := h1.Connect(context.Background(), *releyAddr); err != nil {
		log.Printf("Failed to connect h1 and h2: %v", err)
		return
	}

	// h1 send data to h3
	_, err = h1.NewStream(context.Background(), h3.ID, "/cats")
	if err == nil {
		log.Println("Didnt actually expect to get a stream here. What happened?")
		return
	}

	// Creates a relay address to h3 using h2 as the relay
	relayaddr, err := ma.NewMultiaddr("/p2p/" + releyAddr.ID.Pretty() + "/p2p-circuit/ipfs/" + h3.ID.Pretty())
	if err != nil {
		log.Println(err)
		return
	}

	if err3 != nil {
		fmt.Printf("box start fail %s \n", err)
		return
	}
	h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID)

	h3relayInfo := peer.AddrInfo{
		ID:    h3.ID,
		Addrs: []ma.Multiaddr{relayaddr},
	}
	if err := h1.Connect(context.Background(), h3relayInfo); err != nil {
		log.Printf("Failed to connect h1 and h3: %v", err)
		return
	}

	s, err := h1.NewStream(context.Background(), h3.ID, "/cats")
	if err != nil {
		log.Println("huh, this should have worked: ", err)
		return
	}
	i := []byte("hello im h1 ")
	s.Write(i)
	s.Read(make([]byte, 1)) // block until the handler closes the stream
	select {}
}
