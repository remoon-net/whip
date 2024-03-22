package client

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/shynome/err0/try"
	"github.com/stretchr/testify/assert"
	"remoon.net/link/server"
)

var testEndpoint string

func TestMain(m *testing.M) {
	srv := server.New(0)
	l := try.To1(net.Listen("tcp", "127.0.0.1:0"))
	defer l.Close()
	go http.Serve(l, srv)
	testEndpoint = l.Addr().String()
	m.Run()
}

func TestClient(t *testing.T) {
	const peer = "31ce765283ad48ccf14a827bb4a03e5e2965ce1e6c774a76de09f825b1d08219"
	for _, i := range []string{"1", "2"} {
		func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/x", func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "xxxx"+i)
			})
			mux.HandleFunc("/y", func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "yyyy"+i)
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "zzzz"+i)
			})
			client := New(mux)
			ctx := context.Background()
			sess := try.To1(client.Connect(ctx, "http://"+testEndpoint, peer))
			defer sess.Close()

			for _, c := range []struct {
				path   string
				expect string
			}{
				{"/x", "xxxx" + i},
				{"/y", "yyyy" + i},
				{"/z", "zzzz" + i},
				{"/j", "zzzz" + i},
				{"/", "zzzz" + i},
			} {
				endpoint := "http://" + peer + "@" + testEndpoint + c.path
				resp := try.To1(http.Get(endpoint))
				defer resp.Body.Close()
				body := try.To1(io.ReadAll(resp.Body))
				assert.Equal(t, c.expect, string(body))
			}
		}()
	}
}
