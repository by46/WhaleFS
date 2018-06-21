package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/etcdserver/api/rafthttp"
	"github.com/coreos/etcd/etcdserver/api/snap"
	stats "github.com/coreos/etcd/etcdserver/api/v2stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/spf13/cobra"
)

var (
	raftCmd = &cobra.Command{
		Use: "raft",
		Run: runRaft,
	}
	cluster                 string
	id                      int
	kvport                  int
	join                    bool
	snapshotCatchUpEntriesN uint64 = 10000
)

func init() {
	rootCmd.AddCommand(raftCmd)
	raftCmd.Flags().StringVar(&cluster, "cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	raftCmd.Flags().IntVar(&id, "id", 1, "node ID")
	raftCmd.Flags().IntVar(&kvport, "port", 9121, "key-value server port")
	raftCmd.Flags().BoolVar(&join, "join", false, "join an existing cluster")
}

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

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

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
	log.Printf("replaying WAL of member %d", r.id)
	snapshot := r.loadSnapshot()
	w := r.openWAL(snapshot)
	_, hardState, entries, err := w.ReadAll()
	if err != nil {
		log.Fatalf("raft: failed to read WAL (%v)", err)
	}
	r.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		r.raftStorage.ApplySnapshot(*snapshot)
	}
	r.raftStorage.SetHardState(hardState)

	// append to storage so raft starts at the right place in log
	r.raftStorage.Append(entries)

	// send nil once lastIndex is published so client knows commit channel is current
	if len(entries) > 0 {
		r.lastIndex = entries[len(entries)-1].Index
	} else {
		r.commitC <- nil
	}

	return w
}
func (r *raftNode) loadSnapshot() *raftpb.Snapshot {
	snapshot, err := r.snapshotter.Load()
	if err != nil && err != snap.ErrNoSnapshot {
		log.Fatalf("raft: error loading snapshot (%v)", err)
	}
	return snapshot
}

func (r *raftNode) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(r.waldir) {
		if err := os.Mkdir(r.waldir, 0750); err != nil {
			log.Fatalf("raft: cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(nil, r.waldir, nil)
		if err != nil {
			log.Fatalf("raft: create wal error (%v)", err)
		}
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}

	log.Printf("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	w, err := wal.Open(nil, r.waldir, walsnap)
	if err != nil {
		log.Fatalf("raft: error loading wal (%v)", err)
	}
	return w
}

func (r *raftNode) startRaft() {
	if !fileutil.Exist(r.snapdir) {
		if err := os.Mkdir(r.snapdir, 0750); err != nil {
			log.Fatalf("raft: cannot create dir for snapshot (%v)", err)
		}
	}
	r.snapshotter = snap.New(nil, r.snapdir)
	r.snapshotterReady <- r.snapshotter

	oldwal := wal.Exist(r.waldir)
	r.wal = r.replayWAL()

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

	if oldwal {
		r.node = raft.RestartNode(c)
	} else {
		startPeers := rpeers
		if r.join {
			startPeers = nil
		}
		r.node = raft.StartNode(c, startPeers)
	}

	r.transport = &rafthttp.Transport{
		ID:          types.ID(r.id),
		ClusterID:   0x1000,
		Raft:        r,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(r.id)),
		ErrorC:      make(chan error),
	}
	r.transport.Start()
	for i := range r.peers {
		if i+1 != r.id {
			r.transport.AddPeer(types.ID(i+1), []string{r.peers[i]})
		}
	}
	go r.serveRaft()
	go r.serveChannels()
}
func (r *raftNode) saveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}

	if err := r.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	if err := r.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	return r.wal.ReleaseLockTo(snap.Metadata.Index)
}
func (r *raftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}
	log.Printf("publishing snapshot at index %d", r.snapshotIndex)
	defer log.Printf("finished publishing snapshot at index %d", r.snapshotIndex)

	if snapshotToSave.Metadata.Index <= r.appliedIndex {
		log.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d] + 1", snapshotToSave.Metadata.Index)
	}
	r.commitC <- nil // trigger kvstore to load snapshot

	r.confState = snapshotToSave.Metadata.ConfState
	r.snapshotIndex = snapshotToSave.Metadata.Index
	r.appliedIndex = snapshotToSave.Metadata.Index
}

func (r *raftNode) publishEntries(entries []raftpb.Entry) bool {
	for i := range entries {
		switch entries[i].Type {
		case raftpb.EntryNormal:
			if len(entries[i].Data) == 0 {
				break
			}
			s := string(entries[i].Data)
			select {
			case r.commitC <- &s:
			case <-r.stopc:
				return false
			}
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(entries[i].Data)
			r.confState = *r.node.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					r.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(r.id) {
					log.Println("I've been removed from the cluster! Shutting down.")
					return false
				}
				r.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		r.appliedIndex = entries[i].Index

		if entries[i].Index == r.lastIndex {
			select {
			case r.commitC <- nil:
			case <-r.stopc:
				return false
			}
		}
	}
	return true
}
func (r *raftNode) entriesToApply(entries []raftpb.Entry) (nents []raftpb.Entry) {
	if len(entries) == 0 {
		return
	}
	firstIndex := entries[0].Index
	if firstIndex > r.appliedIndex+1 {
		log.Fatalf("first index of commited entry[%d] should <= progress.appliedIndex[%d]+1", firstIndex, r.appliedIndex)
	}
	if r.appliedIndex-firstIndex+1 < uint64(len(entries)) {
		nents = entries[r.appliedIndex-firstIndex+1:]
	}
	return nents
}
func (r *raftNode) stop() {
	r.stopHTTP()
	close(r.commitC)
	close(r.errorC)
	r.node.Stop()
}
func (r *raftNode) maybeTriggerSnapshot() {
	if r.appliedIndex-r.snapshotIndex <= r.snapCount {
		return
	}
	log.Printf("start snapshot [applied index: %d | last snapshot index: %d]", r.appliedIndex, r.snapshotIndex)

	data, err := r.getSnapshot()
	if err != nil {
		log.Panic(err)
	}
	snap, err := r.raftStorage.CreateSnapshot(r.appliedIndex, &r.confState, data)
	if err != nil {
		panic(err)
	}

	if err := r.saveSnap(snap); err != nil {
		panic(err)
	}

	compactIndex := uint64(1)
	if r.appliedIndex > snapshotCatchUpEntriesN {
		compactIndex = r.appliedIndex - snapshotCatchUpEntriesN
	}
	if err := r.raftStorage.Compact(compactIndex); err != nil {
		panic(err)
	}
	log.Printf("compacted log at index %d", compactIndex)
	r.snapshotIndex = r.appliedIndex
}
func (r *raftNode) writeError(err error) {
	r.stopHTTP()
	close(r.commitC)
	r.errorC <- err
	close(r.errorC)
	r.node.Stop()
}
func (r *raftNode) stopHTTP() {

	r.transport.Stop()
	close(r.httpstopc)
	<-r.httpdonec
}
func (r *raftNode) serveChannels() {
	snap, err := r.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	r.confState = snap.Metadata.ConfState
	r.snapshotIndex = snap.Metadata.Index
	r.appliedIndex = snap.Metadata.Index

	defer r.wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		var confChangeCount uint64 = 0
		for r.proposeC != nil && r.confChangeC != nil {
			select {
			case prop, ok := <-r.proposeC:
				if !ok {
					r.proposeC = nil
				} else {
					r.node.Propose(context.TODO(), []byte(prop))
				}
			case cc, ok := <-r.confChangeC:
				if !ok {
					r.confChangeC = nil
				} else {
					confChangeCount += 1
					cc.ID = confChangeCount
					r.node.ProposeConfChange(context.TODO(), cc)
				}
			}
		}
		close(r.stopc)
	}()

	for {
		select {
		case <-ticker.C:
			r.node.Tick()
		case rd := <-r.node.Ready():
			r.wal.Save(rd.HardState, rd.Entries)
			if !raft.IsEmptySnap(rd.Snapshot) {
				r.saveSnap(rd.Snapshot)
				r.raftStorage.ApplySnapshot(rd.Snapshot)
				r.publishSnapshot(rd.Snapshot)
			}
			r.raftStorage.Append(rd.Entries)
			r.transport.Send(rd.Messages)
			if ok := r.publishEntries(r.entriesToApply(rd.CommittedEntries)); !ok {
				r.stop()
				return
			}
			r.maybeTriggerSnapshot()
			r.node.Advance()
		case err := <-r.transport.ErrorC:
			r.writeError(err)
			return
		case <-r.stopc:
			r.stop()
			return
		}
	}
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
	peers := strings.Split(cluster, ",")
	commitC, errorC, snapshotterReady := newRaftNode(id, peers, join, getSnapshot, proposeC, confChangeC)

	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)

	serveHTTPKVAPI(kvs, kvport, confChangeC, errorC)
}
