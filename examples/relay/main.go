package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	run()
}

func run() {
	fmt.Println("box started")
	server, err := libp2p.New(libp2p.ListenAddrs(), libp2p.EnableRelay())
	if err != nil {
		log.Printf("Failed to create server: %v", err)
		return
	}
	fmt.Println()
	releyAddr, err := peer.AddrInfoFromString("/ip4/192.168.1.23/tcp/7676/ipfs/QmZLwcrA4NLWUehrrh6fjC8xunUd7uv1yUbDxkyNYmdzKZ")

	if err := server.Connect(context.Background(), *releyAddr); err != nil {
		log.Printf("Failed to connect server and h2: %v", err)
		return
	}

	for _, value := range server.Addrs() {
		fmt.Printf("%s/ipfs/%s        %s\n", value, server.ID(), server.ID().Pretty())
	}

	server.SetStreamHandler("/cats", func(s network.Stream) {
		log.Println("Meow! It worked!")
		buf := make([]byte, 1024)
		s.Read(buf)
		fmt.Println(string(buf[:]))
		s.Close()
	})
	select {}

}
