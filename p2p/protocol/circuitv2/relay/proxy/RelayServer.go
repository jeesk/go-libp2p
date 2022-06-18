package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	bhost "github.com/libp2p/go-libp2p/p2p/host/blank"
	msmux "github.com/libp2p/go-libp2p/p2p/muxer/muxer-multistream"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	csms "github.com/libp2p/go-libp2p/p2p/net/conn-security-multistream"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	util "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/util"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
)

type MyFiflter struct {
}

type MyLimit struct {
}

func (fi MyFiflter) AllowReserve(p peer.ID, a ma.Multiaddr) bool {
	fmt.Printf("peerId(%s) AllowReserve \n", p.Pretty())
	return true
}

// AllowConnect returns true if a source peer, with a given multiaddr is allowed to connect
// to a destination peer.
func (fi MyFiflter) AllowConnect(src peer.ID, srcAddr ma.Multiaddr, dest peer.ID) bool {
	fmt.Printf("peerId(%s) AllowConnect  %s \n", src.String(), dest.String())
	return true
}

func main() {

	fmt.Println("relayServer started")
	key, err2 := util.LoadOrCreatePrivateKey("/data/home/song/project/go-libp2p/p2p/protocol/circuitv2/relay/proxy/RelayServer.config")
	if err2 != nil {
		fmt.Print("创建privateKey faild ")
		return
	}

	p, err := peer.IDFromPrivateKey(key)
	if err != nil {
		fmt.Println("get PrivateKey fail")
		return
	}

	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		fmt.Println("create Peerstore faild")
		return
	}
	err = ps.AddPrivKey(p, key)
	if err != nil {
		fmt.Println("addPrivateKey faild")
		return
	}

	bwr := metrics.NewBandwidthCounter()
	netw, err := swarm.NewSwarm(p, ps, swarm.WithMetrics(bwr))
	if err != nil {
		fmt.Println("create swarm faild")
		return
	}
	// 带宽设置

	id := netw.LocalPeer()
	pk := netw.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(insecure.ID, insecure.NewWithIdentity(id, pk))

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)
	upgrader, err := tptu.New(secMuxer, stMuxer)

	tpt, err := tcp.NewTCPTransport(upgrader, nil)
	if err != nil {
		fmt.Println("create tcp transport faild")
		return
	}
	if err := netw.AddTransport(tpt); err != nil {
		fmt.Println("add Transport faild")
		return
	}
	port := "7671"
	ipv4 := "/ip4/0.0.0.0/tcp/" + port
	//ipv6 := "/ip6/::/tcp/" + port
	err = netw.Listen(ma.StringCast(ipv4))

	if err != nil {
		fmt.Println("listen faild")
		return
	}

	relayHost := bhost.NewBlankHost(netw)
	for _, value := range relayHost.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", value.String(), relayHost.ID().Pretty())
	}
	_, err = relay.New(relayHost,
		relay.WithResources(relay.DefaultResources()),
		relay.WithLimit(relay.DefaultLimit()),
		relay.WithACL(&MyFiflter{}))

	// defer r.Close()
	select {}

}
