package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
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
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
)

func main() {

	privk, pubk, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		fmt.Println("创建私钥失败")
		return
	}

	p, err := peer.IDFromPublicKey(pubk)
	if err != nil {
		fmt.Println("获取peer 失败")
		return
	}

	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		fmt.Println("创建PeerStore 失败")
		return
	}
	err = ps.AddPrivKey(p, privk)
	if err != nil {
		fmt.Println("增加私钥 失败")
		return
	}

	bwr := metrics.NewBandwidthCounter()
	netw, err := swarm.NewSwarm(p, ps, swarm.WithMetrics(bwr))
	if err != nil {
		fmt.Println("创建swarm 失败")
		return
	}
	// 带宽设置

	id := netw.LocalPeer()
	pk := netw.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(insecure.ID, insecure.NewWithIdentity(id, pk))

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)
	u, err := tptu.New(secMuxer, stMuxer)

	tpt, err := tcp.NewTCPTransport(u, nil)
	if err != nil {
		fmt.Println("创建swarm 失败")
		return
	}
	if err := netw.AddTransport(tpt); err != nil {
		fmt.Println("创建swarm 失败")
		return
	}

	err = netw.Listen(ma.StringCast("/ip4/0.0.0.0/tcp/7676"))
	if err != nil {
		fmt.Println("创建swarm 失败")
		return
	}

	host := bhost.NewBlankHost(netw)

	for _, value := range host.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", value, peer.Encode(host.ID()))
	}

	_, err = relay.New(host) /*	relay.WithACL(nil),
		relay.WithResources(relayv1.DefaultResources())*/

	if err != nil {
		fmt.Println("创建中继服务器失败")
		return
	}
	select {}

}
