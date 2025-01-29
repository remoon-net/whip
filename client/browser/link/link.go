package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/shynome/err0/try"
	"remoon.net/wslink/client"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	})

	client := client.New(mux)
	ctx := context.Background()
	pk := "f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c"
	sess := try.To1(client.Connect(ctx, fmt.Sprintf("ws://%s@127.0.0.1:2234", pk)))
	<-sess.CloseChan()
}
