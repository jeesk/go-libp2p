package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
)

func main() {

	host, err := libp2p.New(
		libp2p.DisableRelay(),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7676"),
	)
	fmt.Println("relayserver id : %s", host.ID())
	for _, value := range host.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", value, peer.Encode(host.ID()))
	}

	if err != nil {
		fmt.Println("create host fail")
		return
	}
	_, err = relayv1.NewRelay(host)
	if err != nil {
		fmt.Println("create relay server fail ")
		return
	}
	fmt.Println("中继服务器的地址" + host.ID().Pretty())
	select {}

}
