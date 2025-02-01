package tftp

import (
	"fmt"
	"net"
	"time"
)

type client struct {
	conn      *net.UDPConn
	opts      map[option]optVal
	blockSize optValBlocksize
	mode      mode
	localAddr *net.UDPAddr
	destAddr  *net.UDPAddr
	timeout   time.Duration
}

func newClient(destHostname string) *client {
	destAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", destHostname, tftpPortInit))
	if err != nil {
		panic(err)
	}
	c := &client{
		opts:      make(map[option]optVal),
		blockSize: blockSizeDefault,
		mode:      ModeOctet,
		localAddr: &net.UDPAddr{},
		destAddr:  destAddr,
		timeout:   5 * time.Second,
	}

	return c
}

func (c *client) listen() error {
	conn, err := net.ListenUDP("udp", c.localAddr)
	if err != nil {
		return fmt.Errorf("%T.listen: could not create connection: %w", c, err)
	}
	c.conn = conn
	return nil
}

func (c *client) send(b []byte) error {
	c.conn.SetDeadline(time.Now().Add(c.timeout))
	n, err := c.conn.WriteToUDP(b, c.destAddr)
	if err != nil {
		return fmt.Errorf("%T.send: could not send to %s", c, c.destAddr)
	}
	if n < len(b) {
		return fmt.Errorf("%T.send: only wrote %d of %d bytes in buffer", c, n, len(b))
	}
	return nil
}

func (c *client) recv(r int) (map[opCode]any, error) {
	b := make([]byte, r)
	c.conn.SetDeadline(time.Now().Add(c.timeout))
	n, remoteAddr, err := c.conn.ReadFromUDP(b)
	if err != nil {
		return nil, fmt.Errorf("%T.recv: could not read from udp: %w", c, err)
	}

	c.destAddr.Port = remoteAddr.Port

	out := make(map[opCode]any)
	b = b[:n]
	opCode := fromTwoBytes[opCode](b[0:2])
	switch opCode {
	case opACK:
		blockNum, err := readAck(b)
		if err != nil {
			return nil, fmt.Errorf("%T.recv: bad ACK: %w", c, err)
		}
		out[opACK] = blockNum
	case opOACK:
		opts, err := readOack(b)
		if err != nil {
			return nil, fmt.Errorf("%T.recv: bad OACK: %w", c, err)
		}
		c.setOpts(opts)
		out[opOACK] = opts
	case opERROR:
		errCode, errMessage, err := readError(b)
		if err != nil {
			return nil, fmt.Errorf("%T.recv: bad ERROR: %w", c, err)
		}
		return nil, fmt.Errorf("%T.recv: received error code %s with message '%s'", c, errCode, errMessage)
	}

	return out, nil
}

func (c *client) recvOack() error {
	b := make([]byte, 256)
	c.conn.SetDeadline(time.Now().Add(c.timeout))
	n, remoteAddr, err := c.conn.ReadFromUDP(b)
	if err != nil {
		return fmt.Errorf("%T.recvOack: could not read from udp: %w", c, err)
	}

	c.destAddr.Port = remoteAddr.Port

	b = b[:n]
	opCode := fromTwoBytes[opCode](b[0:2])
	switch opCode {
	case opOACK, opACK:
		opts, err := readOack(b)
		if err != nil {
			return fmt.Errorf("%T.recvOack: bad OACK: %w", c, err)
		}
		c.setOpts(opts)
	case opERROR:
		errCode, errMessage, err := readError(b)
		if err != nil {
			return fmt.Errorf("%T.recvOack: bad ERROR: %w", c, err)
		}
		return fmt.Errorf("%T.recv: received error code %s with message '%s'", c, errCode, errMessage)
	}

	return nil
}

func (c *client) setOpts(opts map[option]optVal) {
	c.opts = opts
	for option, value := range c.opts {
		switch option {
		case optionBlockSize:
			c.blockSize = value.(optValBlocksize)
		case optionTimeout:
			c.timeout = time.Duration(value.(optValTimeout)) * time.Second
		}
	}
}
