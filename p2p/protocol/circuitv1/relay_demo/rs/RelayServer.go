package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay_demo/util"
)

func main() {

	key, err := util.LoadOrCreatePrivateKey("/data/home/song/project/go-libp2p/p2p/protocol/circuitv1/relay_demo/rs/RelayServer.config")
	if err != nil {
		fmt.Printf("创建privateKey 失败")
		return
	}

	fmt.Println(" RelayServer started ...")
	host, err := libp2p.New(
		libp2p.Identity(key),
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

	_, err = relayv1.NewRelay(host,
		relayv1.WithACL(MyACLFilter{}),
		relayv1.WithResources(relayv1.DefaultResources()),
	)

	if err != nil {
		fmt.Println("create relay server fail ")
		return
	}

	fmt.Println("中继服务器的地址" + host.ID().Pretty())
	select {}

}

type MyACLFilter struct {
}

type Resources struct {
	// MaxCircuits is the maximum number of active relay connections
	MaxCircuits int

	// MaxCircuitsPerPeer is the maximum number of active relay connections per peer
	MaxCircuitsPerPeer int

	// BufferSize is the buffer size for relaying in each direction
	BufferSize int
}

func (maf MyACLFilter) AllowHop(src, dest peer.ID) bool {
	fmt.Println("src: %s  dest: %s", src.Pretty(), dest.Pretty())
	return true
}
