package config

import (
	"fmt"

	host "github.com/libp2p/go-libp2p-host"
	mux "github.com/libp2p/go-stream-muxer"
	msmux "github.com/whyrusleeping/go-smux-multistream"
)

// MuxC is a stream multiplex transport constructor
type MuxC func(h host.Host) (mux.Transport, error)

// MsMuxC is a tuple containing a multiplex transport constructor and a protocol
// ID.
type MsMuxC struct {
	MuxC
	ID string
}

var muxArgTypes = newArgTypeSet(hostType, networkType, peerIDType, pstoreType)

// 这个是一个适配器
// 实现了Transaport 函数直接返回即可

// MuxerConstructor creates a multiplex constructor from the passed parameter
// using reflection.  、// 传入了一个函数
func MuxerConstructor(m interface{}) (MuxC, error) {
	// Already constructed?
	// 强转成一个接口， 如果这个接口是Transport 的话， 直接返回返回。
	if t, ok := m.(mux.Transport); ok {
		return func(_ host.Host) (mux.Transport, error) {
			return t, nil
		}, nil
	}
	// 如果不是函数的话，
	ctor, err := makeConstructor(m, muxType, muxArgTypes)
	if err != nil {
		return nil, err
	}

	return func(h host.Host) (mux.Transport, error) {
		t, err := ctor(h, nil)
		if err != nil {
			return nil, err
		}
		return t.(mux.Transport), nil
	}, nil
}

// 创建一个多路复用连接器
func makeMuxer(h host.Host, tpts []MsMuxC) (mux.Transport, error) {

	muxMuxer := msmux.NewBlankTransport()
	transportSet := make(map[string]struct{}, len(tpts))
	for _, tptC := range tpts {
		if _, ok := transportSet[tptC.ID]; ok {
			return nil, fmt.Errorf("duplicate muxer transport: %s", tptC.ID)
		}
		transportSet[tptC.ID] = struct{}{}
	}
	for _, tptC := range tpts {
		tpt, err := tptC.MuxC(h)
		if err != nil {
			return nil, err
		}
		muxMuxer.AddTransport(tptC.ID, tpt)
	}
	return muxMuxer, nil
}
