package relay_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	bhost "github.com/libp2p/go-libp2p/p2p/host/blank"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	swarmt "github.com/libp2p/go-libp2p/p2p/net/swarm/testing"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"

	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	ma "github.com/multiformats/go-multiaddr"
)

// https://github.com/libp2p/specs/blob/master/relay/circuit-v2.md
func getNetHosts(t *testing.T, ctx context.Context, n int) (hosts []host.Host, upgraders []transport.Upgrader) {
	for i := 0; i < n; i++ {
		privk, pubk, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
		if err != nil {
			t.Fatal(err)
		}

		p, err := peer.IDFromPublicKey(pubk)
		if err != nil {
			t.Fatal(err)
		}

		ps, err := pstoremem.NewPeerstore()
		if err != nil {
			t.Fatal(err)
		}
		err = ps.AddPrivKey(p, privk)
		if err != nil {
			t.Fatal(err)
		}

		bwr := metrics.NewBandwidthCounter()
		netw, err := swarm.NewSwarm(p, ps, swarm.WithMetrics(bwr))
		if err != nil {
			t.Fatal(err)
		}
		// 带宽设置
		upgrader := swarmt.GenUpgrader(t, netw)
		upgraders = append(upgraders, upgrader)

		tpt, err := tcp.NewTCPTransport(upgrader, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := netw.AddTransport(tpt); err != nil {
			t.Fatal(err)
		}

		err = netw.Listen(ma.StringCast("/ip4/127.0.0.1/tcp/0"))
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			h := bhost.NewBlankHost(netw)

			hosts = append(hosts, h)
		} else {
			h := bhost.NewBlankHost(netw)

			hosts = append(hosts, h)
		}

	}

	return hosts, upgraders
}

func connect(t *testing.T, a, b host.Host) {
	pi := peer.AddrInfo{ID: a.ID(), Addrs: a.Addrs()}
	err := b.Connect(context.Background(), pi)
	if err != nil {
		t.Fatal(err)
	}
}

func addTransport(t *testing.T, h host.Host, upgrader transport.Upgrader) {
	if err := client.AddTransport(h, upgrader); err != nil {
		t.Fatal(err)
	}
}

type MyFiflter struct {
}

type MyLimit struct {
}

func (fi MyFiflter) AllowReserve(p peer.ID, a ma.Multiaddr) bool {
	return true
}

// AllowConnect returns true if a source peer, with a given multiaddr is allowed to connect
// to a destination peer.
func (fi MyFiflter) AllowConnect(src peer.ID, srcAddr ma.Multiaddr, dest peer.ID) bool {
	return true
}

func TestBasicRelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts, upgraders := getNetHosts(t, ctx, 3)
	relayServer := hosts[1]
	boxServer := hosts[0]
	mobileClient := hosts[2]

	// 增加p2p t网络配置
	addTransport(t, boxServer, upgraders[0]) //  add transation
	addTransport(t, mobileClient, upgraders[2])

	rch := make(chan []byte, 1)
	boxServer.SetStreamHandler("test", func(s network.Stream) {
		defer s.Close()
		defer close(rch)
	addTransport(t, hosts[0], upgraders[0])
	addTransport(t, hosts[2], upgraders[2])
	rch := make(chan []byte, 1)
	relayServer := hosts[1]
	boxServer := hosts[0]
	mobileClient := hosts[2]
	fmt.Println("relayServer :" + relayServer.ID().Pretty())
	fmt.Println("boxServer :" + boxServer.ID().Pretty())
	fmt.Println("mobileClient :" + boxServer.ID().Pretty())
	boxServer.SetStreamHandler("test", func(s network.Stream) {
		defer func() {
			s.Close()
			fmt.Println("close")
		}()
		defer func() {
			fmt.Println("close")
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
				t.Fatal(err)
			}
		}
		fmt.Println(string(buf[:nread]))
		rch <- buf[:nread]
	})

	r, err := relay.New(relayServer) /*	relay.WithACL(nil),
		relay.WithResources(relayv1.DefaultResources())*/
	r, err := relay.New(relayServer,
		relay.WithResources(relay.DefaultResources()),
		relay.WithLimit(relay.DefaultLimit()),
		relay.WithACL(&MyFiflter{}))

	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	connect(t, boxServer, relayServer)
	connect(t, relayServer, mobileClient)
	connect(t, mobileClient, relayServer)

	rinfo := relayServer.Peerstore().PeerInfo(relayServer.ID())
	rsvp, err := client.Reserve(ctx, boxServer, rinfo)
	if err != nil {
		t.Fatal(err)
	}

	if rsvp.Voucher == nil {
		t.Fatal("no reservation voucher")
	}

	raddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", relayServer.ID(), boxServer.ID()))
	if err != nil {
		t.Fatal(err)
	}

	err = mobileClient.Connect(ctx, peer.AddrInfo{ID: boxServer.ID(), Addrs: []ma.Multiaddr{raddr}})
	if err != nil {
		t.Fatal(err)
	}
	conns := mobileClient.Network().ConnsToPeer(boxServer.ID())

	conns := mobileClient.Network().ConnsToPeer(boxServer.ID())
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, but got %d", len(conns))
	}
	if !conns[0].Stat().Transient {
		t.Fatal("expected transient connection")
	}

	s, err := mobileClient.NewStream(network.WithUseTransient(ctx, "test"), boxServer.ID(), "test")
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("relay works!")

	s.Write(msg)

	if err != nil {
		t.Fatal(err)
	}
	/*	if nwritten != len(msg) {
		t.Fatalf("expected to write %d bytes, but wrote %d instead", len(msg), nwritten)
	}*/
	s.CloseWrite()

	// confirm recivied msg
	got := <-rch
	if !bytes.Equal(msg, got) {
		t.Fatalf("Wrong echo; expected %s but got %s", string(msg), string(got))
	}
}

func TestRelayLimitTime(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts, upgraders := getNetHosts(t, ctx, 3)
	addTransport(t, hosts[0], upgraders[0])
	addTransport(t, hosts[2], upgraders[2])

	rch := make(chan error, 1)
	hosts[0].SetStreamHandler("test", func(s network.Stream) {
		defer s.Close()
		defer close(rch)

		buf := make([]byte, 1024)
		_, err := s.Read(buf)

		rch <- err
	})

	rc := relay.DefaultResources()
	rc.Limit.Duration = time.Second

	r, err := relay.New(hosts[1], relay.WithResources(rc))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	connect(t, hosts[0], hosts[1])
	connect(t, hosts[1], hosts[2])

	rinfo := hosts[1].Peerstore().PeerInfo(hosts[1].ID())
	_, err = client.Reserve(ctx, hosts[0], rinfo)
	if err != nil {
		t.Fatal(err)
	}

	raddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", hosts[1].ID(), hosts[0].ID()))
	if err != nil {
		t.Fatal(err)
	}

	err = hosts[2].Connect(ctx, peer.AddrInfo{ID: hosts[0].ID(), Addrs: []ma.Multiaddr{raddr}})
	if err != nil {
		t.Fatal(err)
	}

	conns := hosts[2].Network().ConnsToPeer(hosts[0].ID())
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, but got %d", len(conns))
	}
	if !conns[0].Stat().Transient {
		t.Fatal("expected transient connection")
	}

	s, err := hosts[2].NewStream(network.WithUseTransient(ctx, "test"), hosts[0].ID(), "test")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	n, err := s.Write([]byte("should be closed"))
	s.Write([]byte("should be closed2"))
	if n > 0 {
		t.Fatalf("expected to write 0 bytes, wrote %d", n)
	}
	if err != network.ErrReset {
		t.Fatalf("expected reset, but got %s", err)
	}

	err = <-rch
	if err != network.ErrReset {
		t.Fatalf("expected reset, but got %s", err)
	}
}

func TestRelayLimitData(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts, upgraders := getNetHosts(t, ctx, 3)
	addTransport(t, hosts[0], upgraders[0])
	addTransport(t, hosts[2], upgraders[2])

	rch := make(chan int, 1)
	hosts[0].SetStreamHandler("test", func(s network.Stream) {
		defer s.Close()
		defer close(rch)

		buf := make([]byte, 1024)
		for i := 0; i < 3; i++ {
			n, err := s.Read(buf)
			if err != nil {
				t.Fatal(err)
			}
			rch <- n
		}

		n, err := s.Read(buf)
		if err != network.ErrReset {
			t.Fatalf("expected reset but got %s", err)
		}
		rch <- n
	})

	rc := relay.DefaultResources()
	rc.Limit.Duration = time.Second
	rc.Limit.Data = 4096

	r, err := relay.New(hosts[1], relay.WithResources(rc))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	connect(t, hosts[0], hosts[1])
	connect(t, hosts[1], hosts[2])

	rinfo := hosts[1].Peerstore().PeerInfo(hosts[1].ID())
	_, err = client.Reserve(ctx, hosts[0], rinfo)
	if err != nil {
		t.Fatal(err)
	}

	raddr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", hosts[1].ID(), hosts[0].ID()))
	if err != nil {
		t.Fatal(err)
	}

	err = hosts[2].Connect(ctx, peer.AddrInfo{ID: hosts[0].ID(), Addrs: []ma.Multiaddr{raddr}})
	if err != nil {
		t.Fatal(err)
	}

	conns := hosts[2].Network().ConnsToPeer(hosts[0].ID())
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, but got %d", len(conns))
	}
	if !conns[0].Stat().Transient {
		t.Fatal("expected transient connection")
	}

	s, err := hosts[2].NewStream(network.WithUseTransient(ctx, "test"), hosts[0].ID(), "test")
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err = rand.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		n, err := s.Write(buf)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(buf) {
			t.Fatalf("expected to write %d bytes but wrote %d", len(buf), n)
		}

		n = <-rch
		if n != len(buf) {
			t.Fatalf("expected to read %d bytes but read %d", len(buf), n)
		}
	}

	buf = make([]byte, 4096)
	_, err = rand.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	s.Write(buf)

	n := <-rch
	if n != 0 {
		t.Fatalf("expected to read 0 bytes but read %d", n)
	}

}
