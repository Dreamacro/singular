package singular

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

var (
	// Version request header
	Version = []byte("\x01MAGIC")
)

// Conn define a connect
type Conn struct {
	net.Conn
}

// NewConn return a new Conn
func NewConn(conn net.Conn) Conn {
	return Conn{
		Conn: conn,
	}
}

// Send conn send data
func (conn *Conn) Send(serialized []byte) (err error) {
	err = binary.Write(conn, binary.BigEndian, Version)
	if err != nil {
		return err
	}
	err = binary.Write(conn, binary.BigEndian, int32(len(serialized)))
	if err != nil {
		return err
	}
	_, err = conn.Write(serialized)
	return err
}

// Receive conn receive data
func (conn *Conn) Receive() (buf []byte, err error) {
	version := make([]byte, len(Version))
	err = binary.Read(conn, binary.BigEndian, &version)
	if err != nil || !bytes.Equal(version, Version) {
		return version, errors.New("Version not match")
	}
	var msgLength int32
	err = binary.Read(conn, binary.BigEndian, &msgLength)
	if err != nil {
		return
	}

	buf = make([]byte, msgLength)
	err = binary.Read(conn, binary.BigEndian, buf)

	return buf, err
}
