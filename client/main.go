package main

import (
	"log"
	"net"
	"os"

	"github.com/songgao/water"
)

const (
	// Define the MTU for the TUN interface
	mtu = 1500
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <server-ip:port>", os.Args[0])
	}
	serverAddrStr := os.Args[1]

	log.Println("Starting Go VPN client...")

	// 1. Create a new TUN interface
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "vpn-tun0"

	iface, err := water.New(config)
	if err != nil {
		log.Fatalf("Failed to create TUN interface: %v", err)
	}
	log.Printf("TUN interface created: %s", iface.Name())

	// Note: You must configure the IP address and bring the interface up manually.
	// Example for Linux:
	// sudo ip addr add 10.8.0.2/24 dev vpn-tun0
	// sudo ip link set up dev vpn-tun0

	// 2. Connect to the server via UDP
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr)
	if err != nil {
		log.Fatalf("Failed to resolve server address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to server at %s", serverAddr.String())

	// Goroutine to read from UDP and write to TUN
	go func() {
		packet := make([]byte, mtu)
		for {
			n, err := conn.Read(packet)
			if err != nil {
				log.Printf("Error reading from UDP connection: %v", err)
				continue
			}
			// Write packet to the TUN interface
			_, err = iface.Write(packet[:n])
			if err != nil {
				log.Printf("Error writing to TUN interface: %v", err)
			}
		}
	}()

	// 3. Read from TUN and write to UDP
	packet := make([]byte, mtu)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			log.Printf("Error reading from TUN interface: %v", err)
			continue
		}
		// Send the packet to the server
		_, err = conn.Write(packet[:n])
		if err != nil {
			log.Printf("Error writing to UDP connection: %v", err)
		}
	}
}
