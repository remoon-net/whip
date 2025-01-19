package client

import (
	"context"
	"net/http"

	"github.com/hashicorp/yamux"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"nhooyr.io/websocket"
)

type Client struct {
	handler http.Handler
}

var _ http.Handler = (*Client)(nil)

func New(handler http.Handler) *Client {
	return &Client{
		handler: handler,
	}
}

func (c *Client) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.handler.ServeHTTP(w, r)
}

func (c *Client) Connect(ctx context.Context, link string, id string) (sess *yamux.Session, err error) {
	defer err0.Then(&err, nil, nil)
	socket, resp, err := websocket.Dial(ctx, link, &websocket.DialOptions{
		Subprotocols: []string{"link", id},
	})
	if err != nil {
		if resp != nil && (400 <= resp.StatusCode && resp.StatusCode <= 499) {
			return nil, &ServerRejected{err, resp}
		}
		return nil, err
	}
	conn := websocket.NetConn(ctx, socket, websocket.MessageBinary)
	sess = try.To1(yamux.Server(conn, nil))
	go http.Serve(sess, c)
	return sess, nil
}

// ServerRejected Error: Server Response Status Code bewteen [400,499]
type ServerRejected struct {
	error
	Response *http.Response
}
