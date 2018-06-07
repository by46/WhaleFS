package main

import (
	"whalefs/server"
)

func main() {
	srv := server.NewServer()
	srv.ListenAndServe()
}
