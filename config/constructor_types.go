package config

import (
	"fmt"
	"reflect"

	security "github.com/libp2p/go-conn-security"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	pnet "github.com/libp2p/go-libp2p-interface-pnet"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	transport "github.com/libp2p/go-libp2p-transport"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	filter "github.com/libp2p/go-maddr-filter"
	mux "github.com/libp2p/go-stream-muxer"
)

// 构造器配置类

var (
	// interfaces
	hostType      = reflect.TypeOf((*host.Host)(nil)).Elem()
	networkType   = reflect.TypeOf((*inet.Network)(nil)).Elem()
	transportType = reflect.TypeOf((*transport.Transport)(nil)).Elem()
	muxType       = reflect.TypeOf((*mux.Transport)(nil)).Elem()
	securityType  = reflect.TypeOf((*security.Transport)(nil)).Elem()
	protectorType = reflect.TypeOf((*pnet.Protector)(nil)).Elem()
	privKeyType   = reflect.TypeOf((*crypto.PrivKey)(nil)).Elem()
	pubKeyType    = reflect.TypeOf((*crypto.PubKey)(nil)).Elem()
	pstoreType    = reflect.TypeOf((*pstore.Peerstore)(nil)).Elem()

	// concrete types
	peerIDType   = reflect.TypeOf((peer.ID)(""))
	filtersType  = reflect.TypeOf((*filter.Filters)(nil))
	upgraderType = reflect.TypeOf((*tptu.Upgrader)(nil))
)

//Reflect.Type 和函数的映射关系
var argTypes = map[reflect.Type]constructor{
	// 类型和构造器的映射
	upgraderType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u },
	hostType:      func(h host.Host, u *tptu.Upgrader) interface{} { return h },
	networkType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Network() },
	muxType:       func(h host.Host, u *tptu.Upgrader) interface{} { return u.Muxer },
	securityType:  func(h host.Host, u *tptu.Upgrader) interface{} { return u.Secure },
	protectorType: func(h host.Host, u *tptu.Upgrader) interface{} { return u.Protector },
	filtersType:   func(h host.Host, u *tptu.Upgrader) interface{} { return u.Filters },
	peerIDType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.ID() },
	privKeyType:   func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PrivKey(h.ID()) },
	pubKeyType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore().PubKey(h.ID()) },
	pstoreType:    func(h host.Host, u *tptu.Upgrader) interface{} { return h.Peerstore() },
}

// 返回类型和对应的构造器映射
func newArgTypeSet(types ...reflect.Type) map[reflect.Type]constructor {
	result := make(map[reflect.Type]constructor, len(types))

	for _, ty := range types {
		// 通过类型获取到函数，最后设置到返回值中去。
		c, ok := argTypes[ty]
		if !ok {
			panic(fmt.Sprintf("missing constructor for type %s", ty))
		}
		result[ty] = c
	}
	return result
}
