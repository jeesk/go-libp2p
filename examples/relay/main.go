package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/examples/relay/util"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {
	run()
}

func run() {
	key, err := util.LoadOrCreatePrivateKey("/data/home/song/project/go-libp2p/examples/relay/boxServer.config")
	if err != nil {
		fmt.Printf("创建privateKey 失败")
		return
	}

	fmt.Println("box started")
	server, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7677"),
		libp2p.EnableRelay(),
		libp2p.Identity(key),
	)
	if err != nil {
		log.Printf("Failed to create server: %v", err)
		return
	}
	fmt.Println()

	releyAddr, err := peer.AddrInfoFromString("/ip4/148.70.94.33/tcp/7676/p2p/QmXTX2EsvCGU2Z8HKveLgb4uMVGC7WsfzXGGbEvofKJbfv")

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
