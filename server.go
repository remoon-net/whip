package main

import (
	"log"
	"net"
	"net/http"

	"github.com/shynome/err0/try"
	"remoon.net/link/server"
)

func main() {
	srv := server.New()
	l := try.To1(net.Listen("tcp", "127.0.0.1:7799"))
	defer l.Close()
	log.Println("server is running on ", l.Addr().String())
	http.Serve(l, srv)
}
