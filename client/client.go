package client

import (
	"context"
	"errors"
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
	socket, _, err := websocket.Dial(ctx, link, &websocket.DialOptions{
		Subprotocols: []string{"link", id},
	})
	if err != nil {
		return nil, err
	}
	conn := websocket.NetConn(ctx, socket, websocket.MessageBinary)
	sess = try.To1(yamux.Server(conn, nil))
	go http.Serve(sess, c)
	if err = socket.Ping(ctx); err == nil {
		// do nothing
	} else if sess, err := sess.AcceptStream(); err != nil {
		var ce websocket.CloseError
		if errors.As(err, &ce) && ((3400 <= ce.Code && ce.Code <= 3499) || (4400 <= ce.Code && ce.Code <= 4499)) {
			return nil, NewServerRejected(ce)
		}
		return nil, err
	} else {
		sess.Close() // 永远也不会到达这里, 但还是写上这个
	}
	return sess, nil
}

// ServerRejected Error: Server Response Status Code bewteen [400,499]
type ServerRejected struct {
	websocket.CloseError
}

func NewServerRejected(err websocket.CloseError) error {
	return &ServerRejected{
		CloseError: err,
	}
}
