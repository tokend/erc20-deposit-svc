package client

import (
	"github.com/tokend/erc20-deposit-svc/internal/horizon/path"
	"io"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/tokend/keypair"
)

type Interface interface {
	Get(endpoint string) ([]byte, error)
	Put(endpoint string, body io.Reader) ([]byte, error)
	Post(endpoint string, body io.Reader) ([]byte, error)
}

type Client struct {
	throttle <-chan time.Time
	client   *http.Client
	signer   keypair.Full
	resolve  path.Resolver
}

func New(client *http.Client, base *url.URL) *Client {
	return &Client{
		client:   client,
		resolve:  path.NewResolver(base),
		throttle: throttle(),
	}
}

func (c *Client) WithSigner(signer keypair.Full) *Client {
	return &Client{
		client:   c.client,
		signer:   signer,
		resolve:  c.resolve,
		throttle: c.throttle,
	}
}

func (c *Client) Resolve() path.Resolver {
	return c.resolve
}

func throttle() chan time.Time {
	burst := 2 << 10
	ch := make(chan time.Time, burst)

	go func() {
		tick := time.Tick(1 * time.Second)
		// prefill buffer
		for i := 0; i < burst; i++ {
			ch <- time.Now()
		}
		for {
			select {
			case ch <- <-tick:
			default:
			}
		}
	}()
	return ch
}
