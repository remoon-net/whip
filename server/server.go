package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"slices"
	"strings"

	"github.com/docker/go-units"
	"github.com/hashicorp/yamux"
	"github.com/maypok86/otter"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"nhooyr.io/websocket"
)

type Server struct {
	hub otter.Cache[string, *httputil.ReverseProxy]
}

var _ http.Handler = (*Server)(nil)

func New(size int) *Server {
	if size == 0 {
		size = 500 * units.MiB
	}
	hub := try.To1(
		otter.
			MustBuilder[string, *httputil.ReverseProxy](size).
			Build(),
	)
	srv := &Server{
		hub: hub,
	}
	return srv
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	protocols := headerTokens(r.Header, "Sec-Websocket-Protocol")
	if s := indexFunc(protocols, "link"); s != -1 {
		protocols := protocols[s:]
		if len(protocols) != 2 {
			http.Error(w, "unkown which peer", http.StatusBadRequest)
			return
		}
		peer := protocols[1]
		if b, _ := hex.DecodeString(peer); len(b) != 32 {
			http.Error(w, "peer id 不规范", http.StatusBadRequest)
			return
		}
		srv.RegisterHandler(w, r, peer)
		return
	}
	var peer string
	if s := indexFunc(protocols, "peer"); s != -1 {
		protocols := protocols[s:]
		if len(protocols) != 2 {
			http.Error(w, "unkown which peer", http.StatusBadRequest)
			return
		}
		peer = protocols[1]
	} else if c, err := r.Cookie("xhe-peer-id"); err == nil {
		peer = c.Value
	}
	if peer == "" {
		var pwd string
		peer, pwd, _ = r.BasicAuth()
		if peer != "" && pwd == "xhe" {
			http.SetCookie(w, &http.Cookie{
				Name:  "xhe-peer-id",
				Value: peer,
			})
		}
	}
	if peer == "" {
		w.Header().Set("WWW-Authenticate", `Basic`)
		http.Error(w, "unkown which peer", http.StatusUnauthorized)
		return
	}
	srv.linkHandler(w, r, peer)
}

func (srv *Server) RegisterHandler(w http.ResponseWriter, r *http.Request, peer string) (err error) {
	defer err0.Then(&err, nil, func() {
		http.Error(w, err.Error(), 500)
	})
	hub := srv.hub
	if hub.Has(peer) {
		return fmt.Errorf("该地址已被使用")
	}
	socket := try.To1(websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
		Subprotocols:   []string{"link"},
	}))
	ctx := r.Context()
	conn := websocket.NetConn(ctx, socket, websocket.MessageBinary)
	sess := try.To1(yamux.Client(conn, nil))
	defer sess.Close()
	endpoint := fmt.Sprintf("http://yamux.proxy/")
	target := try.To1(url.Parse(endpoint))
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sess.Open()
		},
	}
	ok := hub.SetIfAbsent(peer, proxy)
	if !ok {
		return fmt.Errorf("容量不够了")
	}
	defer hub.Delete(peer)
	<-sess.CloseChan()
	return nil
}

func (srv *Server) linkHandler(w http.ResponseWriter, r *http.Request, peer string) (err error) {
	defer err0.Then(&err, nil, func() {
		http.Error(w, err.Error(), 500)
	})
	hub := srv.hub
	proxy, ok := hub.Get(peer)
	if !ok || proxy == nil {
		return fmt.Errorf("不存在")
	}
	proxy.ServeHTTP(w, r)
	return nil
}

func headerTokens(h http.Header, key string) []string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	var tokens []string
	for _, v := range h[key] {
		v = strings.TrimSpace(v)
		for _, t := range strings.Split(v, ",") {
			t = strings.TrimSpace(t)
			tokens = append(tokens, t)
		}
	}
	return tokens
}

func indexFunc(s []string, t string) int {
	return slices.IndexFunc(s, func(s string) bool { return strings.EqualFold(s, t) })
}
