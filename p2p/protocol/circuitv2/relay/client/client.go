package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec/insecure"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	bhost "github.com/libp2p/go-libp2p/p2p/host/blank"
	msmux "github.com/libp2p/go-libp2p/p2p/muxer/muxer-multistream"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	csms "github.com/libp2p/go-libp2p/p2p/net/conn-security-multistream"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	tptu "github.com/libp2p/go-libp2p/p2p/net/upgrader"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/util"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
	"io/ioutil"
)

func main() {

	fmt.Println("mobile_client started")

	key, err2 := util.LoadOrCreatePrivateKey("/data/home/song/project/go-libp2p/p2p/protocol/circuitv2/relay/client/client.config")
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
	port := "7669"
	ipv4 := "/ip4/0.0.0.0/tcp/" + port
	ipv6 := "/ip6/::/tcp/" + port
	err = netw.Listen(ma.StringCast(ipv4), ma.StringCast(ipv6))
	if err != nil {
		fmt.Println("listen faild")
		return
	}

	mobileClient := bhost.NewBlankHost(netw)
	for _, value := range mobileClient.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", value.String(), mobileClient.ID().Pretty())
	}

	if err := client.AddTransport(mobileClient, upgrader); err != nil {
		fmt.Println("addTransport  faild")
		return
	}

	relayServerAddr, err2 := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/7671/ipfs/QmaWomNyyYh9TbGkTLnPMBrSyKNZQbncxEyqH8sqgTLxn1")
	if err2 != nil {
		fmt.Println("relayServer addr error")
		return
	}

	err2 = mobileClient.Connect(context.Background(), *relayServerAddr)
	if err2 != nil {
		fmt.Println("连接中继服务器失败")
		return
	}

	boxServer, err2 := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/7672/ipfs/QmdREM4pPNUupeDqtYBqGeAEhEWtr1D4GGEsXF4Aw8w8oT")
	if err2 != nil {
		fmt.Println("boxServer addr error")
		return
	}

	sprintf := fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", relayServerAddr.ID.Pretty(), boxServer.ID.Pretty())
	raddr, err := ma.NewMultiaddr(sprintf)
	if err != nil {
		fmt.Print("解析地址失败")
		return
	}
	background := context.Background()
	err = mobileClient.Connect(background,
		peer.AddrInfo{ID: boxServer.ID, Addrs: []ma.Multiaddr{raddr}})
	if err != nil {
		fmt.Printf("连接box Server 失败 err: %v", err)
		return
	}

	conns := mobileClient.Network().ConnsToPeer(boxServer.ID)
	if len(conns) != 1 {
		fmt.Printf("expected 1 connection, but got %d", len(conns))
		return
	}
	if !conns[0].Stat().Transient {
		fmt.Print("expected transient connection")
		return
	}

	s, err := mobileClient.NewStream(network.WithUseTransient(background, "test"), boxServer.ID, "test")
	if err != nil {
		fmt.Println("连接失败")
	}

	//	msg := []byte("relay works!")
	file, err2 := ioutil.ReadFile("/data/home/song/project/go-libp2p/p2p/protocol/circuitv2/relay/client/boot.txt")
	if err2 != nil {
		fmt.Println("read file faild ")
		return
	}

	nwritten, err := s.Write(file)
	if err != nil {
		fmt.Println("写入消息失败")
		return
	}
	if nwritten != len(file) {
		fmt.Printf("expected to write %d bytes, but wrote %d instead", len(file), nwritten)
	}
	bytes := make([]byte, 50)
	read, err2 := s.Read(bytes)
	fmt.Println("接收到server 端的消息: ", string(bytes[:read]))
	s.CloseWrite()

	select {}
}
