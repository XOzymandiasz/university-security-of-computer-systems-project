package serverclient

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) HealthCheck() string {
	return "successfully health check"
}
