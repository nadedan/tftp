package tftp

import (
	"fmt"
	"io"
	"net"
	"time"
)

const (
	ModeNetascii mode = "netascii" // ASCII text mode
	ModeOctet    mode = "octet"    // Raw binary mode
)

// Put executes a TFTP PUT.
//
// Parameters:
//   - toHostname: a string representing the hostname/ip of the server that we want to PUT to
//   - atFilename: a string of file we are writing to at the destination
//   - data: an io.Reader where we will get the data to send
//   - opts: options for the PUT
//     -- WithBlocksize
//     -- WithTimeoutSeconds
//     -- WithMode
//     -- FromLocalInterface
func Put(toHostname string, atFilename string, data io.Reader, opts ...putOptFunc) error {

	client := newClient(toHostname)
	for _, opt := range opts {
		opt(client)
	}

	err := client.listen()
	if err != nil {
		return fmt.Errorf("Put: could not get client to listen: %w", err)
	}

	tftpOpts := map[option]optVal{
		optionBlockSize: client.blockSize,
		optionTimeout:   optValTimeout(client.timeout.Seconds()),
	}
	err = client.send(wrqPacket(atFilename, client.mode, tftpOpts))
	if err != nil {
		return fmt.Errorf("tftp.Put: failed to send WRQ: %w", err)
	}

	err = client.recvOack()
	if err != nil {
		return fmt.Errorf("tftp.Put: failed to receive ACK for WRQ: %w", err)
	}

	block := make([]byte, client.blockSize)
	done := false
	for blkNum := 1; !done; blkNum++ {
		n, readErr := data.Read(block)
		if n < int(client.blockSize) {
			done = true
		}

		sendErr := client.send(dataPacket(blockNum(blkNum), block[:n]))
		if readErr != nil {
			if readErr != io.EOF {
				return fmt.Errorf("tftp.Put: failed to get new block %d to send: %w", blkNum, readErr)
			}
			done = true
		}
		if sendErr != nil {
			return fmt.Errorf("tftp.Put: failed to send block %d: %w", blkNum, sendErr)
		}

		_, err = client.recv(maxRxSize)
		if err != nil {
			return fmt.Errorf("tftp.Put: failed to receive ACK for block %d: %w", blkNum, err)
		}
	}

	return nil
}

// WithBlocksize allows us to attempt to set the block size of a PUT.
//
// The default block size is 512 bytes. We will request to use the set
// blocksize. The server will do one of three things:
//   - agree and the PUT will proceed with our set block size
//   - tell us to use something different and we will obey
//   - abort the transfer with an error code and we will be quiet
func WithBlocksize(size uint16) putOptFunc {
	return func(c *client) {
		c.blockSize = optValBlocksize(size)
	}
}

// WithTimeoutSeconds allows us to attemp to set the packet timeout of a PUT.
// Timeout is given in uint8 seconds.
//
// We will request to use the set timeout. The server will do one of three things:
//   - agree and the PUT will proceed with our set timeout
//   - tell us to use something different and we will obey
//   - abort the transfer with an error code and we will be quiet
func WithTimeoutSeconds(seconds uint8) putOptFunc {
	return func(c *client) {
		c.timeout = time.Duration(seconds) * time.Second
	}
}

// WithMode sets the mode of the transfer.
//
// Available modes are:
//   - ModeOctet: (default) raw binary transfer
//   - ModeNetascii: send data in ascii
func WithMode(m mode) putOptFunc {
	return func(c *client) {
		c.mode = m
	}
}

// FromLocalInterface specifies the IP address of the
// interface that we want to communicate from.
func FromLocalInterface(interfaceIp string) putOptFunc {
	addr, err := net.ResolveUDPAddr("udp", interfaceIp)
	if err != nil {
		panic(fmt.Sprintf("bad local interface ip %s: %s", interfaceIp, err))
	}
	return func(c *client) {
		c.localAddr = addr
	}
}

type putOptFunc func(*client)
