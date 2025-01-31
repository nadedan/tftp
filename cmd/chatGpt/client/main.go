package main

import (
	"fmt"
	"net"
	"time"
)

const (
	tftpPort = 69
	opWRQ    = 2
	opDATA   = 3
	opACK    = 4
	mode     = "octet"
	timeout  = 5 * time.Second
)

// makeWRQPacket creates a TFTP Write Request (WRQ) packet.
func makeWRQPacket(filename string) []byte {
	packet := make([]byte, 2+len(filename)+1+len(mode)+1)
	packet[0], packet[1] = 0, opWRQ
	copy(packet[2:], filename)
	packet[2+len(filename)] = 0
	copy(packet[3+len(filename):], mode)
	packet[len(packet)-1] = 0
	return packet
}

// makeDataPacket creates a TFTP DATA packet.
func makeDataPacket(blockNum uint16, data []byte) []byte {
	packet := make([]byte, 4+len(data))
	packet[0], packet[1] = 0, opDATA
	packet[2], packet[3] = byte(blockNum>>8), byte(blockNum&0xFF)
	copy(packet[4:], data)
	return packet
}

// sendTFTPPut performs a TFTP zero-byte PUT operation.
func sendTFTPPut(serverIP, filename string) error {
	serverAddr := fmt.Sprintf("%s:%d", serverIP, tftpPort)
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer conn.Close()

	// Send WRQ
	wrqPacket := makeWRQPacket(filename)
	if _, err = conn.Write(wrqPacket); err != nil {
		return fmt.Errorf("failed to send WRQ: %v", err)
	}

	// Receive ACK (Block 0)
	conn.SetReadDeadline(time.Now().Add(timeout))
	ack := make([]byte, 4)
	n, err := conn.Read(ack)
	if err != nil || n < 4 || ack[1] != opACK {
		return fmt.Errorf("failed to receive ACK for WRQ: %v", err)
	}

	// Extract server's chosen UDP port (may be different from 69)
	serverPort := fmt.Sprintf("%s:%d", serverIP, int(ack[2])<<8|int(ack[3]))
	conn, err = net.Dial("udp", serverPort)
	if err != nil {
		return fmt.Errorf("failed to connect to TFTP data port: %v", err)
	}
	defer conn.Close()

	// Send empty data packet (Block 1)
	dataPacket := makeDataPacket(1, []byte{})
	if _, err = conn.Write(dataPacket); err != nil {
		return fmt.Errorf("failed to send zero-byte data packet: %v", err)
	}

	// Receive ACK (Block 1)
	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err = conn.Read(ack)
	if err != nil || n < 4 || ack[1] != opACK {
		return fmt.Errorf("failed to receive final ACK: %v", err)
	}

	fmt.Println("Zero-byte file successfully uploaded.")
	return nil
}

func main() {
	serverIP := "localhost"
	filename := "/RUNGEM/0x00000000"

	if err := sendTFTPPut(serverIP, filename); err != nil {
		fmt.Println("Error:", err)
	}
}
