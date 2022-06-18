package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/util"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"io"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	ctx := context.Background()
	fmt.Println("box_server started")
	key, err2 := util.LoadOrCreatePrivateKey("/data/home/song/project/go-libp2p/p2p/protocol/circuitv2/relay/box_server/box_server.config")
	port := "7672"
	ipv4 := "/ip4/0.0.0.0/tcp/" + port
	ipv6 := "/ip6/::/tcp/" + port

	options := make([]config.Option, 10)
	options = append(options,
		libp2p.EnableHolePunching(),
		libp2p.ListenAddrStrings(ipv4, ipv6),
		libp2p.NATPortMap(),
		libp2p.Identity(key),
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
		),
	)
	boxServerHost, err2 := libp2p.New(options...)

	if err2 != nil {
		fmt.Println("listen faild")
		return
	}

	/*	if err := client.AddTransport(boxServerHost, upgrader); err != nil {
		fmt.Println("addTransport  faild")
		return
	}*/

	rch := make(chan []byte, 1)
	boxServerHost.SetStreamHandler("test", func(s network.Stream) {
		defer func() {
			s.Close()
		}()
		defer func() {
			close(rch)
		}()

		// box 收到消息
		buf := make([]byte, 1024)
		nread := 0
		for nread < len(buf) {
			n, err := s.Read(buf[nread:])
			nread += n
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("process data faild")
			}
		}
		file, err2 := ioutil.ReadFile("/data/home/song/project/go-libp2p/p2p/protocol/circuitv2/relay/client/boot.txt")
		if err2 != nil {
			fmt.Println("read file faild ")
			return
		}

		s.Write(file)
		fmt.Printf("box server 收到数据 %s \n", string(buf[:nread]))
		rch <- buf[:nread]
	})

	relayServerAddr, err2 := peer.AddrInfoFromString("/ip4/148.70.94.33/tcp/7676/p2p/QmXTX2EsvCGU2Z8HKveLgb4uMVGC7WsfzXGGbEvofKJbfv")
	//relayServerAddr, err2 := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/7671/ipfs/QmaWomNyyYh9TbGkTLnPMBrSyKNZQbncxEyqH8sqgTLxn1\n")

	if err2 != nil {
		fmt.Printf("get Addr faild %s \n", err2)
		return
	}

	err2 = boxServerHost.Connect(context.Background(), *relayServerAddr)
	if err2 != nil {
		fmt.Println("连接失败", err2)
		return
	}

	rsvp, err := client.Reserve(ctx, boxServerHost, *relayServerAddr)
	if err != nil {
		fmt.Println("reserve 失败")
		return
	}

	if rsvp.Voucher == nil {
		fmt.Println("没有预订的凭证")
		return
	}

	boxServerHost.SetStreamHandler("/cats", func(s network.Stream) {
		log.Println("Meow! It worked!")
		buf := make([]byte, 1024)
		s.Read(buf)
		fmt.Println("boxServer收到消息" + string(buf[:]))
		s.Close()
	})
	for _, value := range boxServerHost.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", value.String(), boxServerHost.ID().Pretty())
	}

	timer := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-timer.C:
			conns := boxServerHost.Network().ConnsToPeer(relayServerAddr.ID)
			if len(conns) == 0 {
				fmt.Printf("expected 1 connection, but got %d \n", len(conns))
				err2 = boxServerHost.Connect(context.Background(), *relayServerAddr)
				if err2 != nil {
					fmt.Println("连接失败", err2)
					continue
				}
				rsvp, err := client.Reserve(ctx, boxServerHost, *relayServerAddr)
				if err != nil {
					fmt.Println("reserve 失败")
					continue
				}

				if rsvp.Voucher == nil {
					fmt.Println("没有预订的凭证")
					continue
				}
			} else {
				if !conns[0].Stat().Transient {
					fmt.Println("连接relay serer 成功")
				}
			}
		}
	}

}
