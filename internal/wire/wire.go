package wire

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

func Send(conn net.Conn, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	if _, err = conn.Write(lenBuf); err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

func Receive(conn net.Conn, out any) error {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return err
	}

	n := binary.BigEndian.Uint32(lenBuf)
	data := make([]byte, n)

	if _, err := io.ReadFull(conn, data); err != nil {
		return err
	}

	return json.Unmarshal(data, out)
}

// SendBinaryFrame sends a binary frame directly without JSON/base64 encoding
// Format: [4 bytes: frame length][frame data]
func SendBinaryFrame(conn net.Conn, frameData []byte) error {
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(frameData)))

	if _, err := conn.Write(lenBuf); err != nil {
		return err
	}
	_, err := conn.Write(frameData)
	return err
}

// ReceiveBinaryFrame receives a binary frame directly
func ReceiveBinaryFrame(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}

	n := binary.BigEndian.Uint32(lenBuf)
	if n > 50*1024*1024 { // Max 50MB frame size
		return nil, io.ErrUnexpectedEOF
	}

	data := make([]byte, n)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}