package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/docker/go-units"
	"github.com/shynome/err0/try"
	"remoon.net/link/server"
)

var args struct {
	size string
}

func init() {
	flag.StringVar(&args.size, "size", "500M", "最大容纳量")
}

func main() {
	flag.Parse()
	size := try.To1(units.FromHumanSize(args.size))
	srv := server.New(int(size))
	l := try.To1(net.Listen("tcp", "127.0.0.1:7799"))
	defer l.Close()
	log.Println("server is running on ", l.Addr().String())
	http.Serve(l, srv)
}
