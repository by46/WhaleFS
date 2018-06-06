package main

import (
	"WhaleFS/server"
)

func main() {
	srv := server.NewServer()
	srv.ListenAndServe()
}
