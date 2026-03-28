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

// GetConn returns the underlying connection (for binary frame reading)
func (c *Client) GetConn() net.Conn {
	return c.conn
}

func connect(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
	}
	return conn, nil
}
