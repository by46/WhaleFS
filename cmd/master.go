package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/chrislusf/raft"
	"github.com/labstack/echo"
	"github.com/spf13/cobra"
)

var (
	masterCmd = &cobra.Command{
		Use:   "master",
		Short: "master server",
		Run:   runMaster,
	}
)

func init() {
	rootCmd.AddCommand(masterCmd)
}

type RaftServer struct {
	raftServer raft.Server
	app        *echo.Echo
}

func (r *RaftServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.app.Any(pattern, func(ctx echo.Context) error {
		handler(ctx.Response(), ctx.Request())
		return nil
	})
}

func (r *RaftServer) join(ctx echo.Context) error {
	fmt.Printf("Processing incoming join. Current Leader %s, Self %s, Peers %v", r.raftServer.Leader(),
		r.raftServer.Name(),
		r.raftServer.Peers())

	text, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	command := &raft.DefaultJoinCommand{}
	if err := json.Unmarshal(text, command); err != nil {
		return err
	}
	return nil
}

func (r *RaftServer) status(ctx echo.Context) error {
	return nil
}

func runMaster(cmd *cobra.Command, args []string) {
	var err error
	server := &RaftServer{
		app: echo.New(),
	}

	transporter := raft.NewHTTPTransporter("/cluster", 0)
	transporter.Transport.MaxIdleConnsPerHost = 1024

	server.raftServer, err = raft.NewServer("localhost:9003", "fs/master", transporter, nil, nil, "")
	if err != nil {
		os.Exit(1)
	}
	transporter.Install(server.raftServer, server)
	server.raftServer.SetHeartbeatInterval(500 * time.Millisecond)
	server.raftServer.SetElectionTimeout(1000 * time.Millisecond)
	server.raftServer.Start()

	server.app.POST("/cluster/join", server.join)
	server.app.GET("/cluster/status", server.status)

	if err := server.app.Start(":9003"); err != nil {
		fmt.Printf("start failed %v", err)
	}
}
