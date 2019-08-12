package net

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
)

// Client tcp client
type Client struct {
	conn   net.Conn
	Server *Server
}

// Send text message to client
func (c *Client) Send(data []byte, code int16) error {
	packer := new(Packer)
	packer.Code = code
	packer.Length = int32(len(data))
	packer.Msg = data

	bytes, e := packer.PackToByte()
	if e != nil {
		return e
	}
	_, e1 := c.conn.Write(bytes)
	return e1
}

// Conn get net.Conn
func (c *Client) Conn() net.Conn {
	return c.conn
}

// Close close tcp
func (c *Client) Close() error {
	return c.conn.Close()
}

// Read client data from channel
func (c *Client) work() {
	if c.Server.onNewClient != nil {
		c.Server.onNewClient(c)
	}

	reader := bufio.NewReader(c.conn)
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		lenData := len(data)
		if atEOF && lenData == 0 {
			return 0, nil, nil
		}
		if lenData > 10 {
			indexSplit := bytes.IndexAny(data, SPLITER)
			if indexSplit >= 0 {
				length := int32(0)
				binary.Read(bytes.NewReader(data[6:10]), binary.LittleEndian, &length)

				indexNext := indexSplit + 4 + 2 + 4 + int(length)
				if indexNext <= lenData {
					return indexNext, data[:indexNext], nil
				}
			}
		}

		if atEOF {
			return lenData, data, nil
		}
		return 0, nil, nil
	})
	for scanner.Scan() {
		scannedPack := new(Packer)
		scannedPack.Unpack(bytes.NewReader(scanner.Bytes()))
		// log.Println(scannedPack)
		if c.Server.onNewMessage != nil {
			c.Server.onNewMessage(c, scannedPack.Code, string(scannedPack.Msg))
		}
	}

	if c.Server.onClientClosed != nil {
		c.Server.onClientClosed(c, nil)
	}
}
