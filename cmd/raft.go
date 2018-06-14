package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/coreos/etcd/etcdserver/api/rafthttp"
	"github.com/coreos/etcd/etcdserver/api/snap"
	"github.com/coreos/etcd/etcdserver/api/v2stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
)

var (
	raftCmd = &cobra.Command{
		Use: "raft",
		Run: runRaft,
	}
	cluster string
	id      int
	kvport  int
	join    bool
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
	proposeC    <-chan string            // proposed messages (k,v)
	confChangeC <-chan raftpb.ConfChange // proposed cluster config changes
	commitC     chan<- *string           // entries committed to log (k,v)
	errorC      chan<- error             // errors from raft session

	id          int
	join        bool
	peers       []string
	getSnapshot func() ([]byte, error)
	waldir      string
	snapdir     string
	lastIndex   uint64

	node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	snapCount uint64

	transport *rafthttp.Transport

	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown
	httpdonec chan struct{} // signals http server shutdown complete
}

func newRaftNode(id int,
	peers []string,
	join bool,
	getSnapshot func() ([]byte, error),
	proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange) (<-chan *string, <-chan error, <-chan *snap.Snapshotter) {

	commitC := make(chan *string)
	errorC := make(chan error)

	r := &raftNode{
		proposeC:         proposeC,
		confChangeC:      confChangeC,
		commitC:          commitC,
		errorC:           errorC,
		id:               id,
		peers:            peers,
		join:             join,
		waldir:           fmt.Sprintf("raft-%d", id),
		snapdir:          fmt.Sprintf("raft-%d-snap", id),
		getSnapshot:      getSnapshot,
		snapCount:        1000,
		stopc:            make(chan struct{}),
		httpstopc:        make(chan struct{}),
		httpdonec:        make(chan struct{}),
		snapshotterReady: make(chan *snap.Snapshotter, 1),
	}
	go r.startRaft()
	return commitC, errorC, r.snapshotterReady
}

func (r *raftNode) replayWAL() *wal.WAL {
	log.Printf("replaying WAL for member %d", r.id)
	return nil
}
func (r *raftNode) startRaft() {
	if !fileutil.Exist(r.snapdir) {
		if err := os.Mkdir(r.snapdir, 0750); err != nil {
			log.Fatalf("raft: cannot create dir for snapshot (%v)", err)
		}
	}
	r.raftStorage = raft.NewMemoryStorage()

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

type httpKVAPI struct {
	store       *kvstore
	confChangeC chan<- raftpb.ConfChange
}

func (h *httpKVAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.RequestURI
	switch r.Method {
	case "PUT":
		v, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read on PUT(%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}
		h.store.Propose(key, string(v))
		w.WriteHeader(http.StatusNoContent)
	case "GET":
		v, exists := h.store.Lookup(key)
		if !exists {
			http.Error(w, "Failed to GET", http.StatusNoContent)
			return
		}
		w.Write([]byte(v))

	case "POST":
		url, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read on POST (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		nodeId, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			log.Printf("Failed to convert ID for conf change (%v)\n", err)
			http.Error(w, "Failed on POST", http.StatusBadRequest)
			return
		}

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  nodeId,
			Context: url,
		}
		h.confChangeC <- cc
		w.WriteHeader(http.StatusNoContent)

	case "DELETE":
		nodeId, err := strconv.ParseUint(key[1:], 0, 64)
		if err != nil {
			log.Printf("Failed to convert ID for conf change (%v)\n", err)
			http.Error(w, "Failed on DELETE", http.StatusBadRequest)
			return
		}
		cc := raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: nodeId,
		}
		h.confChangeC <- cc
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "PUT")
		w.Header().Add("Allow", "GET")
		w.Header().Add("Allow", "POST")
		w.Header().Add("Allow", "DELETE")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func init() {
	rootCmd.AddCommand(raftCmd)
	raftCmd.Flags().StringVar(&cluster, "cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	raftCmd.Flags().IntVar(&id, "id", 1, "node ID")
	raftCmd.Flags().IntVar(&kvport, "port", 9121, "key-value server port")
	raftCmd.Flags().BoolVar(&join, "join", false, "join an existing cluster")
}

func serveHTTPKVAPI(kv *kvstore, port int, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	srv := http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: &httpKVAPI{
			store:       kv,
			confChangeC: confChangeC,
		},
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}
func runRaft(cmd *cobra.Command, args []string) {
	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)
	var kvs *kvstore
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(1, []string{}, false, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	serveHTTPKVAPI(kvs, kvport, confChangeC, errorC)
}
