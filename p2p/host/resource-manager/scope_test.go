package rcmgr

import (
	"testing"

	"github.com/libp2p/go-libp2p-core/network"
)

func checkResources(t *testing.T, rc *resources, st network.ScopeStat) {
	t.Helper()

	if rc.nconnsIn != st.NumConnsInbound {
		t.Fatalf("expected %d inbound conns, got %d", st.NumConnsInbound, rc.nconnsIn)
	}
	if rc.nconnsOut != st.NumConnsOutbound {
		t.Fatalf("expected %d outbound conns, got %d", st.NumConnsOutbound, rc.nconnsOut)
	}
	if rc.nstreamsIn != st.NumStreamsInbound {
		t.Fatalf("expected %d inbound streams, got %d", st.NumStreamsInbound, rc.nstreamsIn)
	}
	if rc.nstreamsOut != st.NumStreamsOutbound {
		t.Fatalf("expected %d outbound streams, got %d", st.NumStreamsOutbound, rc.nstreamsOut)
	}
	if rc.nfd != st.NumFD {
		t.Fatalf("expected %d file descriptors, got %d", st.NumFD, rc.nfd)
	}
	if rc.memory != st.Memory {
		t.Fatalf("expected %d reserved bytes of memory, got %d", st.Memory, rc.memory)
	}
}

func TestResources(t *testing.T) {
	rc := newResources(&StaticLimit{
		Memory:          4096,
		StreamsInbound:  1,
		StreamsOutbound: 1,
		ConnsInbound:    1,
		ConnsOutbound:   1,
		FD:              1,
	})

	checkResources(t, rc, network.ScopeStat{})

	if err := rc.checkMemory(1024); err != nil {
		t.Fatal(err)
	}
	if err := rc.checkMemory(4096); err != nil {
		t.Fatal(err)
	}
	if err := rc.checkMemory(8192); err == nil {
		t.Fatal("expected memory check to fail")
	}

	if err := rc.reserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{Memory: 1024})

	if err := rc.reserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{Memory: 2048})

	if err := rc.reserveMemory(4096); err == nil {
		t.Fatal("expected memory reservation to fail")
	}
	checkResources(t, rc, network.ScopeStat{Memory: 2048})

	rc.releaseMemory(1024)
	checkResources(t, rc, network.ScopeStat{Memory: 1024})

	if err := rc.reserveMemory(2048); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{Memory: 3072})

	rc.releaseMemory(3072)
	checkResources(t, rc, network.ScopeStat{})

	if err := rc.addStream(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{NumStreamsInbound: 1})

	if err := rc.addStream(network.DirInbound); err == nil {
		t.Fatal("expected addStream to fail")
	}
	checkResources(t, rc, network.ScopeStat{NumStreamsInbound: 1})

	rc.removeStream(network.DirInbound)
	checkResources(t, rc, network.ScopeStat{})

	if err := rc.addStream(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{NumStreamsOutbound: 1})

	if err := rc.addStream(network.DirOutbound); err == nil {
		t.Fatal("expected addStream to fail")
	}
	checkResources(t, rc, network.ScopeStat{NumStreamsOutbound: 1})

	rc.removeStream(network.DirOutbound)
	checkResources(t, rc, network.ScopeStat{})

	if err := rc.addConn(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{NumConnsInbound: 1})

	if err := rc.addConn(network.DirInbound); err == nil {
		t.Fatal("expected addConn to fail")
	}
	checkResources(t, rc, network.ScopeStat{NumConnsInbound: 1})

	rc.removeConn(network.DirInbound)
	checkResources(t, rc, network.ScopeStat{})

	if err := rc.addConn(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{NumConnsOutbound: 1})

	if err := rc.addConn(network.DirOutbound); err == nil {
		t.Fatal("expected addConn to fail")
	}
	checkResources(t, rc, network.ScopeStat{NumConnsOutbound: 1})

	rc.removeConn(network.DirOutbound)
	checkResources(t, rc, network.ScopeStat{})

	if err := rc.addFD(1); err != nil {
		t.Fatal(err)
	}
	checkResources(t, rc, network.ScopeStat{NumFD: 1})

	if err := rc.addFD(1); err == nil {
		t.Fatal("expected addFD to fail")
	}
	checkResources(t, rc, network.ScopeStat{NumFD: 1})

	rc.removeFD(1)
	checkResources(t, rc, network.ScopeStat{})
}

func TestResourceScopeSimple(t *testing.T) {
	s := NewResourceScope(
		&StaticLimit{
			Memory:          4096,
			StreamsInbound:  1,
			StreamsOutbound: 1,
			ConnsInbound:    1,
			ConnsOutbound:   1,
			FD:              1,
		},
		nil,
	)

	s.IncRef()
	if s.refCnt != 1 {
		t.Fatal("expected refcnt of 1")
	}
	s.DecRef()
	if s.refCnt != 0 {
		t.Fatal("expected refcnt of 0")
	}

	testResourceScopeBasic(t, s)
}

func testResourceScopeBasic(t *testing.T, s *ResourceScope) {
	if err := s.ReserveMemory(2048); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{Memory: 2048})

	if err := s.ReserveMemory(2048); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{Memory: 4096})

	if err := s.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	checkResources(t, s.rc, network.ScopeStat{Memory: 4096})

	s.ReleaseMemory(4096)
	checkResources(t, s.rc, network.ScopeStat{})

	if err := s.AddStream(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{NumStreamsInbound: 1})

	if err := s.AddStream(network.DirInbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	checkResources(t, s.rc, network.ScopeStat{NumStreamsInbound: 1})

	s.RemoveStream(network.DirInbound)
	checkResources(t, s.rc, network.ScopeStat{})

	if err := s.AddStream(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{NumStreamsOutbound: 1})

	if err := s.AddStream(network.DirOutbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	checkResources(t, s.rc, network.ScopeStat{NumStreamsOutbound: 1})

	s.RemoveStream(network.DirOutbound)
	checkResources(t, s.rc, network.ScopeStat{})

	if err := s.AddConn(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{NumConnsInbound: 1})

	if err := s.AddConn(network.DirInbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	checkResources(t, s.rc, network.ScopeStat{NumConnsInbound: 1})

	s.RemoveConn(network.DirInbound)
	checkResources(t, s.rc, network.ScopeStat{})

	if err := s.AddConn(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{NumConnsOutbound: 1})

	if err := s.AddConn(network.DirOutbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	checkResources(t, s.rc, network.ScopeStat{NumConnsOutbound: 1})

	s.RemoveConn(network.DirOutbound)
	checkResources(t, s.rc, network.ScopeStat{})

	if err := s.AddFD(1); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s.rc, network.ScopeStat{NumFD: 1})

	if err := s.AddFD(1); err == nil {
		t.Fatal("expected AddFD to fail")
	}
	checkResources(t, s.rc, network.ScopeStat{NumFD: 1})

	s.RemoveFD(1)
	checkResources(t, s.rc, network.ScopeStat{})
}

func TestResourceScopeTxnBasic(t *testing.T) {
	s := NewResourceScope(
		&StaticLimit{
			Memory:          4096,
			StreamsInbound:  1,
			StreamsOutbound: 1,
			ConnsInbound:    1,
			ConnsOutbound:   1,
			FD:              1,
		},
		nil,
	)

	txn, err := s.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	testResourceScopeBasic(t, txn.(*ResourceScope))
	checkResources(t, s.rc, network.ScopeStat{})

	// check constraint propagation
	if err := txn.ReserveMemory(4096); err != nil {
		t.Fatal(err)
	}
	checkResources(t, txn.(*ResourceScope).rc, network.ScopeStat{Memory: 4096})
	checkResources(t, s.rc, network.ScopeStat{Memory: 4096})
	txn.Done()
	checkResources(t, s.rc, network.ScopeStat{})
	txn.Done() // idempotent
	checkResources(t, s.rc, network.ScopeStat{})
}

func TestResourceScopeTxnZombie(t *testing.T) {
	s := NewResourceScope(
		&StaticLimit{
			Memory:          4096,
			StreamsInbound:  1,
			StreamsOutbound: 1,
			ConnsInbound:    1,
			ConnsOutbound:   1,
			FD:              1,
		},
		nil,
	)

	txn1, err := s.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn2, err := txn1.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	if err := txn2.ReserveMemory(4096); err != nil {
		t.Fatal(err)
	}
	checkResources(t, txn2.(*ResourceScope).rc, network.ScopeStat{Memory: 4096})
	checkResources(t, txn1.(*ResourceScope).rc, network.ScopeStat{Memory: 4096})
	checkResources(t, s.rc, network.ScopeStat{Memory: 4096})

	txn1.Done()
	checkResources(t, s.rc, network.ScopeStat{})
	if err := txn2.ReserveMemory(4096); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}

	txn2.Done()
	checkResources(t, s.rc, network.ScopeStat{})
}

func TestResourceScopeTxnTree(t *testing.T) {
	s := NewResourceScope(
		&StaticLimit{
			Memory:          4096,
			StreamsInbound:  1,
			StreamsOutbound: 1,
			ConnsInbound:    1,
			ConnsOutbound:   1,
			FD:              1,
		},
		nil,
	)

	txn1, err := s.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn2, err := txn1.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn3, err := txn1.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn4, err := txn2.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn5, err := txn2.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	if err := txn3.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, txn3.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn1.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s.rc, network.ScopeStat{Memory: 1024})

	if err := txn4.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, txn4.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn3.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn2.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn1.(*ResourceScope).rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s.rc, network.ScopeStat{Memory: 2048})

	if err := txn5.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, txn5.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn4.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn3.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn2.(*ResourceScope).rc, network.ScopeStat{Memory: 2048})
	checkResources(t, txn1.(*ResourceScope).rc, network.ScopeStat{Memory: 3072})
	checkResources(t, s.rc, network.ScopeStat{Memory: 3072})

	if err := txn1.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, txn5.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn4.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn3.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn2.(*ResourceScope).rc, network.ScopeStat{Memory: 2048})
	checkResources(t, txn1.(*ResourceScope).rc, network.ScopeStat{Memory: 4096})
	checkResources(t, s.rc, network.ScopeStat{Memory: 4096})

	if err := txn5.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	if err := txn4.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	if err := txn3.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	if err := txn2.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	checkResources(t, txn5.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn4.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn3.(*ResourceScope).rc, network.ScopeStat{Memory: 1024})
	checkResources(t, txn2.(*ResourceScope).rc, network.ScopeStat{Memory: 2048})
	checkResources(t, txn1.(*ResourceScope).rc, network.ScopeStat{Memory: 4096})
	checkResources(t, s.rc, network.ScopeStat{Memory: 4096})

	txn1.Done()
	checkResources(t, s.rc, network.ScopeStat{})
}

func TestResourceScopeDAG(t *testing.T) {
	// A small DAG of scopes
	// s1
	// +---> s2
	//        +------------> s5
	//        +----
	// +---> s3 +.  \
	//          | \  -----+-> s4 (a diamond!)
	//          |  ------/
	//          \
	//           ------> s6
	s1 := NewResourceScope(
		&StaticLimit{
			Memory:          4096,
			StreamsInbound:  4,
			StreamsOutbound: 4,
			ConnsInbound:    4,
			ConnsOutbound:   4,
			FD:              4,
		},
		nil,
	)
	s2 := NewResourceScope(
		&StaticLimit{
			Memory:          2048,
			StreamsInbound:  2,
			StreamsOutbound: 2,
			ConnsInbound:    2,
			ConnsOutbound:   2,
			FD:              2,
		},
		[]*ResourceScope{s1},
	)
	s3 := NewResourceScope(
		&StaticLimit{
			Memory:          2048,
			StreamsInbound:  2,
			StreamsOutbound: 2,
			ConnsInbound:    2,
			ConnsOutbound:   2,
			FD:              2,
		},
		[]*ResourceScope{s1},
	)
	s4 := NewResourceScope(
		&StaticLimit{
			Memory:          2048,
			StreamsInbound:  2,
			StreamsOutbound: 2,
			ConnsInbound:    2,
			ConnsOutbound:   2,
			FD:              2,
		},
		[]*ResourceScope{s2, s3, s1},
	)
	s5 := NewResourceScope(
		&StaticLimit{
			Memory:          2048,
			StreamsInbound:  2,
			StreamsOutbound: 2,
			ConnsInbound:    2,
			ConnsOutbound:   2,
			FD:              2,
		},
		[]*ResourceScope{s2, s1},
	)
	s6 := NewResourceScope(
		&StaticLimit{
			Memory:          2048,
			StreamsInbound:  2,
			StreamsOutbound: 2,
			ConnsInbound:    2,
			ConnsOutbound:   2,
			FD:              2,
		},
		[]*ResourceScope{s3, s1},
	)

	if err := s4.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 1024})

	if err := s5.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 2048})

	if err := s6.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 3072})

	if err := s4.ReserveMemory(1024); err == nil {
		t.Fatal("expcted ReserveMemory to fail")
	}
	if err := s5.ReserveMemory(1024); err == nil {
		t.Fatal("expcted ReserveMemory to fail")
	}
	if err := s6.ReserveMemory(1024); err == nil {
		t.Fatal("expcted ReserveMemory to fail")
	}

	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 3072})

	s4.ReleaseMemory(1024)
	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 2048})

	s5.ReleaseMemory(1024)
	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 1024})

	s6.ReleaseMemory(1024)
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})

	if err := s4.AddStream(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsInbound: 1})

	if err := s5.AddStream(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsInbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsInbound: 2})

	if err := s6.AddStream(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsInbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsInbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsInbound: 3})

	if err := s4.AddStream(network.DirInbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	if err := s5.AddStream(network.DirInbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	if err := s6.AddStream(network.DirInbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsInbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsInbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsInbound: 3})

	s4.RemoveStream(network.DirInbound)
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsInbound: 2})

	s5.RemoveStream(network.DirInbound)
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsInbound: 1})

	s6.RemoveStream(network.DirInbound)
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})

	if err := s4.AddStream(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsOutbound: 1})

	if err := s5.AddStream(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsOutbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsOutbound: 2})

	if err := s6.AddStream(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsOutbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsOutbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsOutbound: 3})

	if err := s4.AddStream(network.DirOutbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	if err := s5.AddStream(network.DirOutbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	if err := s6.AddStream(network.DirOutbound); err == nil {
		t.Fatal("expected AddStream to fail")
	}
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsOutbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsOutbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsOutbound: 3})

	s4.RemoveStream(network.DirOutbound)
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsOutbound: 2})

	s5.RemoveStream(network.DirOutbound)
	checkResources(t, s6.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumStreamsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{NumStreamsOutbound: 1})

	s6.RemoveStream(network.DirOutbound)
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})

	if err := s4.AddConn(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsInbound: 1})

	if err := s5.AddConn(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsInbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsInbound: 2})

	if err := s6.AddConn(network.DirInbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsInbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsInbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsInbound: 3})

	if err := s4.AddConn(network.DirInbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	if err := s5.AddConn(network.DirInbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	if err := s6.AddConn(network.DirInbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	checkResources(t, s6.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsInbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsInbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsInbound: 3})

	s4.RemoveConn(network.DirInbound)
	checkResources(t, s6.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsInbound: 2})

	s5.RemoveConn(network.DirInbound)
	checkResources(t, s6.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsInbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsInbound: 1})

	s6.RemoveConn(network.DirInbound)
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})

	if err := s4.AddConn(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsOutbound: 1})

	if err := s5.AddConn(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsOutbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsOutbound: 2})

	if err := s6.AddConn(network.DirOutbound); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsOutbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsOutbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsOutbound: 3})

	if err := s4.AddConn(network.DirOutbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	if err := s5.AddConn(network.DirOutbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	if err := s6.AddConn(network.DirOutbound); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	checkResources(t, s6.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsOutbound: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsOutbound: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsOutbound: 3})

	s4.RemoveConn(network.DirOutbound)
	checkResources(t, s6.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsOutbound: 2})

	s5.RemoveConn(network.DirOutbound)
	checkResources(t, s6.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumConnsOutbound: 1})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{NumConnsOutbound: 1})

	s6.RemoveConn(network.DirOutbound)
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})

	if err := s4.AddFD(1); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumFD: 1})

	if err := s5.AddFD(1); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumFD: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumFD: 2})

	if err := s6.AddFD(1); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumFD: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumFD: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumFD: 3})

	if err := s4.AddFD(1); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	if err := s5.AddFD(1); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	if err := s6.AddFD(1); err == nil {
		t.Fatal("expected AddConn to fail")
	}
	checkResources(t, s6.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s4.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s3.rc, network.ScopeStat{NumFD: 2})
	checkResources(t, s2.rc, network.ScopeStat{NumFD: 2})
	checkResources(t, s1.rc, network.ScopeStat{NumFD: 3})

	s4.RemoveFD(1)
	checkResources(t, s6.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s5.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s2.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s1.rc, network.ScopeStat{NumFD: 2})

	s5.RemoveFD(1)
	checkResources(t, s6.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{NumFD: 1})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{NumFD: 1})

	s6.RemoveFD(1)
	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})
}

func TestResourceScopeDAGTxn(t *testing.T) {
	// A small DAG of scopes
	// s1
	// +---> s2
	//        +------------> s5
	//        +----
	// +---> s3 +.  \
	//          | \  -----+-> s4 (a diamond!)
	//          |  ------/
	//          \
	//           ------> s6
	s1 := NewResourceScope(
		&StaticLimit{
			Memory: 8192,
		},
		nil,
	)
	s2 := NewResourceScope(
		&StaticLimit{
			Memory: 4096 + 2048,
		},
		[]*ResourceScope{s1},
	)
	s3 := NewResourceScope(
		&StaticLimit{
			Memory: 4096 + 2048,
		},
		[]*ResourceScope{s1},
	)
	s4 := NewResourceScope(
		&StaticLimit{
			Memory: 4096 + 1024,
		},
		[]*ResourceScope{s2, s3, s1},
	)
	s5 := NewResourceScope(
		&StaticLimit{
			Memory: 4096 + 1024,
		},
		[]*ResourceScope{s2, s1},
	)
	s6 := NewResourceScope(
		&StaticLimit{
			Memory: 4096 + 1024,
		},
		[]*ResourceScope{s3, s1},
	)

	txn4, err := s4.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn5, err := s5.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	txn6, err := s6.BeginTransaction()
	if err != nil {
		t.Fatal(err)
	}

	if err := txn4.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 1024})

	if err := txn5.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 2048})

	if err := txn6.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 3072})

	if err := txn4.ReserveMemory(4096); err != nil {
		t.Fatal(err)
	}
	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024 + 4096})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 2048 + 4096})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048 + 4096})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 3072 + 4096})

	if err := txn4.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	if err := txn5.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	if err := txn6.ReserveMemory(1024); err == nil {
		t.Fatal("expected ReserveMemory to fail")
	}
	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{Memory: 1024 + 4096})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 2048 + 4096})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048 + 4096})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 3072 + 4096})

	txn4.Done()

	checkResources(t, s6.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 1024})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 2048})

	if err := txn5.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}
	if err := txn6.ReserveMemory(1024); err != nil {
		t.Fatal(err)
	}

	checkResources(t, s6.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s5.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s2.rc, network.ScopeStat{Memory: 2048})
	checkResources(t, s1.rc, network.ScopeStat{Memory: 4096})

	txn5.Done()
	txn6.Done()

	checkResources(t, s6.rc, network.ScopeStat{})
	checkResources(t, s5.rc, network.ScopeStat{})
	checkResources(t, s4.rc, network.ScopeStat{})
	checkResources(t, s3.rc, network.ScopeStat{})
	checkResources(t, s2.rc, network.ScopeStat{})
	checkResources(t, s1.rc, network.ScopeStat{})
}
