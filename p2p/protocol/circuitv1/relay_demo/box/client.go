package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"log"
)

func main() {

	h1, err := libp2p.New(libp2p.EnableRelay())
	if err != nil {
		log.Printf("Failed to create h1: %v", err)
		return
	}
	releyAddr, err := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/7676/ipfs/QmeTUUSTShkaUWK5CXjBM2QogBawUXE543vZAzV8NhHEYa")
	h3, err3 := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/7677/ipfs/QmeUvHkwWHojmxUx5DzZmxjpvSEEGktrz9rCAownvkJd5R")

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
	// err := ioutil.ReadFile("/data/home/song/project/go-libp2p/p2p/protocol/circuitv1/relay_demo/box/《这个修行世界不太正常》.txt")

	/*	i := []byte("hello im h1 ")
		s.Write(i)*/
	file, err := ioutil.ReadFile("/data/home/song/project/go-libp2p/p2p/protocol/circuitv1/relay_demo/box/《这个修行世界不太正常》.txt")
	if err != nil {
		fmt.Println("read file fail")
		return
	}
	s.Write(file)
	s.Read(make([]byte, 1)) // block until the handler closes the stream
	//
}
