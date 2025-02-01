package tftp

import (
	"fmt"
	"io"
	"net"
	"time"
)

const maxRxSize = 1024

func Put(toHostname string, atFilename string, data io.Reader, opts ...putOptFunc) error {

	client := newClient(toHostname)
	for _, opt := range opts {
		opt(client)
	}

	err := client.listen()
	if err != nil {
		return fmt.Errorf("Put: could not get client to listen: %w", err)
	}

	tftpOpts := map[Option]optVal{
		OptionBlockSize: client.blockSize,
		OptionTimeout:   optValTimeout(client.timeout.Seconds()),
	}
	fmt.Printf("sending wrq to %s\n", client.destAddr)
	err = client.send(Wrq(atFilename, client.mode, tftpOpts))
	if err != nil {
		return fmt.Errorf("tftp.Put: failed to send WRQ: %w", err)
	}

	fmt.Printf("waiting for oack\n")
	err = client.recvOack()
	if err != nil {
		return fmt.Errorf("tftp.Put: failed to receive ACK for WRQ: %w", err)
	}

	block := make([]byte, client.blockSize)
	done := false
	for blockNum := 1; !done; blockNum++ {
		fmt.Printf("block %d\n", blockNum)
		n, readErr := data.Read(block)
		if n < int(client.blockSize) {
			done = true
		}

		sendErr := client.send(Data(BlockNum(blockNum), block[:n]))
		if readErr != nil {
			if readErr != io.EOF {
				return fmt.Errorf("tftp.Put: failed to get new block %d to send: %w", blockNum, readErr)
			}
			done = true
		}
		if sendErr != nil {
			return fmt.Errorf("tftp.Put: failed to send block %d: %w", blockNum, sendErr)
		}

		fmt.Printf("waiting for ack\n")
		_, err = client.recv(maxRxSize)
		if err != nil {
			return fmt.Errorf("tftp.Put: failed to receive ACK for block %d: %w", blockNum, err)
		}
	}

	return nil
}

type putOptFunc func(*client)

func WithBlocksize(size uint16) putOptFunc {
	return func(c *client) {
		c.blockSize = optValBlocksize(size)
	}
}

func WithTimeoutSeconds(seconds uint8) putOptFunc {
	return func(c *client) {
		c.timeout = time.Duration(seconds) * time.Second
	}
}

func WithMode(mode Mode) putOptFunc {
	return func(c *client) {
		c.mode = mode
	}
}

func FromLocalIp(ip net.IP) putOptFunc {
	return func(c *client) {
		c.localAddr = &net.UDPAddr{IP: ip}
	}
}
