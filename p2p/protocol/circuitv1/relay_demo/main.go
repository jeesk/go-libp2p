package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay_demo/util"
	"log"
)

func main() {
	run()
}

func run() {
	key, err := util.LoadOrCreatePrivateKey("/data/home/song/project/go-libp2p/p2p/protocol/circuitv1/relay_demo/boxServer.config")
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
	releyAddr, err := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/7676/ipfs/QmeTUUSTShkaUWK5CXjBM2QogBawUXE543vZAzV8NhHEYa")

	if err := server.Connect(context.Background(), *releyAddr); err != nil {
		log.Printf("Failed to connect server and h2: %v", err)
		return
	}

	for _, value := range server.Addrs() {
		fmt.Printf("%s/ipfs/%s        %s\n", value, server.ID(), server.ID().Pretty())
	}

	server.SetStreamHandler("/cats", func(s network.Stream) {
		log.Println("Meow! It worked!")
		buf := make([]byte, 2048)
		count := 0
		for true {
			read, err2 := s.Read(buf)
			fmt.Println("read 读取 %d 字节", count)
			if err2 != nil {
				fmt.Printf("read error ")
				return
			}
			count = count + read
			if read == -1 {
				fmt.Println("一共读取 %d 字节", read)
				break
			}
		}

	})
	select {}

}
