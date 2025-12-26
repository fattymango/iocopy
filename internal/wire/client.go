package wire

import (
	"net"
)

type Client struct {
	conn net.Conn
}

func NewClient(ip, port string) (*Client, error) {
	conn, err := connect(net.JoinHostPort(ip, port))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

// Read reads a message from the connection
func (c *Client) Read() (*Message, error) {
	var msg Message
	if err := Receive(c.conn, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Write sends a message to the connection
func (c *Client) Write(msg *Message) error {
	return Send(c.conn, msg)
}

// Close closes the connection
func (c *Client) Close() error {
	return c.conn.Close()
}

func connect(addr string) (net.Conn, error) {
	return net.Dial("tcp", addr)
}
