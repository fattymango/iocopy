package wire

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	addr   string
	ln     net.Listener
	conn   net.Conn
	onConn func(*Server, net.Conn)
}

func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// Check if it's a port already in use error
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Err.Error() == "bind: Only one usage of each socket address (protocol/network address/port) is normally permitted." || 
			   opErr.Err.Error() == "bind: address already in use" {
				return nil, fmt.Errorf("port %s is already in use. Please close the existing instance or use a different port with -port flag", addr)
			}
		}
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	return &Server{
		addr: addr,
		ln:   ln,
	}, nil
}

// Start starts the server and accepts connections
func (s *Server) Start(onConn func(*Server, net.Conn)) error {
	s.onConn = onConn
	go func() {
		for {
			conn, err := s.ln.Accept()
			if err != nil {
				log.Printf("[server] Accept error: %v", err)
				continue
			}
			remoteAddr := conn.RemoteAddr()
			log.Printf("[server] New connection accepted from %s", remoteAddr)

			// Handle connection immediately in a separate goroutine to avoid blocking
			if s.onConn != nil {
				// Call handler in a goroutine to ensure it doesn't block accepting new connections
				go func(c net.Conn) {
					s.onConn(s, c)
				}(conn)
			} else {
				conn.Close()
			}
		}
	}()
	log.Printf("[server] Listening on %s", s.addr)
	return nil
}

// SetConn sets the active connection for this server instance
func (s *Server) SetConn(conn net.Conn) {
	s.conn = conn
}

// Read reads a message from the active connection
func (s *Server) Read() (*Message, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("no active connection")
	}
	var msg Message
	if err := Receive(s.conn, &msg); err != nil {
		if err == io.EOF {
			log.Printf("[server] Connection closed by peer")
		}
		return nil, err
	}
	return &msg, nil
}

// Write sends a message to the active connection
func (s *Server) Write(msg *Message) error {
	if s.conn == nil {
		return fmt.Errorf("no active connection")
	}
	return Send(s.conn, msg)
}

// Close closes the active connection and listener
func (s *Server) Close() error {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}
