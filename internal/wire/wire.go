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
