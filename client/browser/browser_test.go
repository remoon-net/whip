package browser

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/shynome/err0/try"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	buildTry()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	caddy := exec.CommandContext(ctx, "caddy", "file-server", "--listen", "127.0.0.1:6111")
	try.To(caddy.Start())
	server := exec.CommandContext(ctx, "wslink", "serve", "--listen", ":2234")
	try.To(server.Start())
	m.Run()
}

func TestWASM(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	{
		ctx, _ := chromedp.NewContext(ctx)
		err := chromedp.Run(ctx,
			chromedp.Navigate("http://127.0.0.1:6111/wasm_exec.html"),
		)
		try.To(err)
	}
	time.Sleep(2 * time.Second)
	link := "http://f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c@127.0.0.1:2234"
	resp := try.To1(http.Get(link))
	body := try.To1(io.ReadAll(resp.Body))
	assert.Equal(t, "ok", string(body))
}

func buildTry() {
	cmd := exec.Command("go", "build", "-o", "client.wasm", "./link")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stderr = os.Stderr
	try.To(cmd.Run())
}
