package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	tftpPort = 69
	opWRQ    = 2
	opDATA   = 3
	opACK    = 4
	timeout  = 5 * time.Second
)

func newConn(port int) *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return nil
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Failed to start UDP server:", err)
		return nil
	}
	return conn
}

// sendACK sends an ACK packet for a given block number.
func sendACK(conn *net.UDPConn, addr *net.UDPAddr, blockNum uint16) error {
	ack := []byte{0, opACK, byte(blockNum >> 8), byte(blockNum & 0xFF)}
	_, err := conn.WriteToUDP(ack, addr)
	return err
}

// handleWRQ handles an incoming WRQ (Write Request).
func handleWRQ(conn *net.UDPConn, addr *net.UDPAddr, request []byte) {

	// Extract filename
	parts := bytes.Split(request[2:], []byte{0})
	if len(parts) < 2 {
		fmt.Println("Invalid WRQ packet")
		return
	}
	filename := string(parts[0])
	fmt.Printf("Received WRQ for file: %s\n", filename)

	port := 8000 + rand.Int31n(1000)
	// Respond with ACK for block 0
	if err := sendACK(conn, addr, uint16(port)); err != nil {
		fmt.Println("Failed to send ACK for WRQ:", err)
		return
	}

	conn = newConn(int(port))
	go func() {
		// Wait for the first (empty) data packet
		conn.SetReadDeadline(time.Now().Add(timeout))
		buffer := make([]byte, 516)
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil || n < 4 || buffer[1] != opDATA {
			fmt.Println("Did not receive expected DATA packet:", err)
			return
		}

		// Check if it's an empty data packet (zero-byte file)
		blockNum := uint16(buffer[2])<<8 | uint16(buffer[3])
		if n == 4 {
			fmt.Println("Zero-byte file received successfully!")
		} else {
			fmt.Println("Received non-empty data packet, ignoring extra content.")
		}

		// Send final ACK
		if err := sendACK(conn, remoteAddr, blockNum); err != nil {
			fmt.Println("Failed to send final ACK:", err)
		}
	}()
}

// startTFTPServer starts a minimal TFTP server.
func startTFTPServer() {
	conn := newConn(tftpPort)
	defer conn.Close()
	fmt.Println("TFTP Server started on port", tftpPort)

	buffer := make([]byte, 516)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Failed to read UDP packet:", err)
			continue
		}

		// Determine if it's a WRQ request
		if n >= 2 && buffer[1] == opWRQ {
			handleWRQ(conn, remoteAddr, buffer[:n])
		}
	}
}

func main() {
	startTFTPServer()
}
