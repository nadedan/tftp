package tftp

import (
	"fmt"
	"net"
	"time"
)

type client struct {
	conn      *net.UDPConn
	opts      map[Option]OptVal
	blockSize OptValBlocksize
	mode      Mode
	destAddr  *net.UDPAddr
	timeout   time.Duration
}

func newClient(localAddr *net.UDPAddr, destHostname string) (*client, error) {
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("newClient: could not create connection for new client: %w", err)
	}

	c := &client{
		conn:      conn,
		opts:      make(map[Option]OptVal),
		blockSize: BlockSizeDefault,
		mode:      ModeOctet,
		destAddr: &net.UDPAddr{
			IP:   net.IP(destHostname),
			Port: TftpPortInit,
		},
		timeout: 5 * time.Second,
	}

	return c, nil
}

func (c *client) send(b []byte, port int) error {
	n, err := c.conn.WriteToUDP(b, c.destAddr)
	if err != nil {
		return fmt.Errorf("%T.send: could not send to %s", c, c.destAddr)
	}
	if n < len(b) {
		return fmt.Errorf("%T.send: only wrote %d of %d bytes in buffer", c, n, len(b))
	}

	return nil
}

func (c *client) recv(r int) (map[OpCode]any, error) {
	b := make([]byte, r)
	n, remoteAddr, err := c.conn.ReadFromUDP(b)
	if err != nil {
		return nil, fmt.Errorf("%T.recv: could not read from udp: %w", c, err)
	}

	c.destAddr.Port = remoteAddr.Port

	out := make(map[OpCode]any)
	b = b[:n]
	opCode := fromTwoBytes[OpCode](b[0:2])
	switch opCode {
	case OpACK:
		blockNum, err := readAck(b)
		if err != nil {
			return nil, fmt.Errorf("%T.recv: bad ACK: %w", c, err)
		}
		out[OpACK] = blockNum
	case OpOACK:
		opts, err := readOack(b)
		if err != nil {
			return nil, fmt.Errorf("%T.recv: bad OACK: %w", c, err)
		}
		c.setOpts(opts)
	case OpERROR:
		errCode, errMessage, err := readError(b)
		if err != nil {
			return nil, fmt.Errorf("%T.recv: bad ERROR: %w", c, err)
		}
		return nil, fmt.Errorf("%T.recv: received error code %s with message '%s'", c, errCode, errMessage)
	}

	return nil, nil
}

func (c *client) setOpts(opts map[Option]OptVal) {
	c.opts = opts
	for option, value := range c.opts {
		switch option {
		case OptionBlockSize:
			c.blockSize = value.(OptValBlocksize)
		case OptionTimeout:
			c.timeout = time.Duration(value.(OptValTimeout)) * time.Second
		}
	}
}
