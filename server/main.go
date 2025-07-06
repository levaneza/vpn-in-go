package main

import (
	"log"
	"net"

	"github.com/songgao/water"
)

const (
	// Define the port the server will listen on
	listenPort = 8888
	// Define the MTU for the TUN interface
	mtu = 1500
)

func main() {
	log.Println("Starting Go VPN server...")

	// 1. Create a new TUN interface
	// The name "tun0" is a suggestion, the OS will pick one if not available
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
	// sudo ip addr add 10.8.0.1/24 dev vpn-tun0
	// sudo ip link set up dev vpn-tun0

	// 2. Listen for incoming UDP packets
	addr := net.UDPAddr{
		Port: listenPort,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port %d: %v", listenPort, err)
	}
	defer conn.Close()
	log.Printf("Listening for clients on UDP port %d", listenPort)

	// We need to know the client's address to send packets back
	var clientAddr *net.UDPAddr

	// Goroutine to read from TUN and write to UDP
	go func() {
		packet := make([]byte, mtu)
		for {
			n, err := iface.Read(packet)
			if err != nil {
				log.Printf("Error reading from TUN interface: %v", err)
				continue
			}
			if clientAddr != nil {
				// Send the packet from TUN to the client
				_, err = conn.WriteToUDP(packet[:n], clientAddr)
				if err != nil {
					log.Printf("Error writing to UDP: %v", err)
				}
			}
		}
	}()

	// 3. Read from UDP and write to TUN
	packet := make([]byte, mtu)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(packet)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		// The first packet from a client sets its address
		if clientAddr == nil || !clientAddr.IP.Equal(remoteAddr.IP) || clientAddr.Port != remoteAddr.Port {
			log.Printf("New client connected: %s", remoteAddr)
			clientAddr = remoteAddr
		}

		// Write the packet from UDP to the TUN interface
		_, err = iface.Write(packet[:n])
		if err != nil {
			log.Printf("Error writing to TUN interface: %v", err)
		}
	}
}
