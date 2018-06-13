package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/coreos/etcd/etcdserver/api/rafthttp"
	"github.com/coreos/etcd/etcdserver/api/v2stats"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
)

var (
	raftCmd = &cobra.Command{
		Use: "raft",
		Run: runRaft,
	}
)

type stoppableListener struct {
	*net.TCPListener
	stopc <-chan struct{}
}

func newStoppableListener(addr string, stopc <-chan struct{}) (*stoppableListener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &stoppableListener{
		ln.(*net.TCPListener),
		stopc,
	}, nil
}

func (l stoppableListener) Accept() (c net.Conn, err error) {
	connc := make(chan *net.TCPConn, 1)
	errc := make(chan error, 1)
	go func() {
		tc, err := l.AcceptTCP()
		if err != nil {
			errc <- err
			return
		}
		connc <- tc
	}()
	select {
	case <-l.stopc:
		return nil, errors.New("server stopped")
	case err := <-errc:
		return nil, err
	case tc := <-connc:
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
		return tc, nil
	}
}

type raftNode struct {
	id    int
	join  bool
	peers []string

	node        raft.Node
	raftStorage *raft.MemoryStorage

	transport *rafthttp.Transport

	httpstopc chan struct{}
}

func (r *raftNode) startRaft() {
	rpeers := make([]raft.Peer, len(r.peers))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}

	c := &raft.Config{
		ID:              uint64(r.id),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         r.raftStorage,
		MaxSizePerMsg:   1 << 20,
		MaxInflightMsgs: 256,
	}
	if r.join == false {
		r.node = raft.RestartNode(c)
	} else {
		r.node = raft.StartNode(c, rpeers)
	}
	r.transport = &rafthttp.Transport{
		ID:          types.ID(r.id),
		ClusterID:   0x1000,
		Raft:        r,
		ServerStats: v2stats.NewServerStats("", ""),
		LeaderStats: v2stats.NewLeaderStats(strconv.Itoa(r.id)),
		ErrorC:      make(chan error),
	}
	r.transport.Start()
	for i := range r.peers {
		if i+1 != r.id {
			r.transport.AddPeer(types.ID(i+1), []string{r.peers[i]})
		}
	}
	go r.serveRaft()
}
func (r *raftNode) serveRaft() {
	url, err := url.Parse(r.peers[r.id-1])
	if err != nil {
		log.Fatalf("raft: Failed parsing URL (%v)", err)
	}
	ln, err := newStoppableListener(url.Host, r.httpstopc)
	if err != nil {
		log.Fatalf("raft: Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{
		Handler: r.transport.Handler(),
	}).Serve(ln)
	select {
	case <-r.httpstopc:
	default:
		log.Fatalf("raft: Failed to serve rafthttp (%v)", err)
	}
	close(r.httpstopc)
}

func (r *raftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}
func (r *raftNode) ReportUnreachable(id uint64)                          {}
func (r *raftNode) IsIDRemoved(id uint64) bool                           { return false }
func (r *raftNode) Process(ctx context.Context, m raftpb.Message) error {
	return r.node.Step(ctx, m)
}

func init() {
	rootCmd.AddCommand(raftCmd)
}

func runRaft(cmd *cobra.Command, args []string) {
}
