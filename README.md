# Go VPN

A simple Layer 3 VPN application written in Go. This project demonstrates how to create a basic VPN using TUN devices for tunneling IP packets between a client and a server over a UDP connection.

**Disclaimer:** This is a proof-of-concept and is **NOT SECURE**. The traffic is sent in plaintext without any encryption or authentication. Do not use this in a production environment.

## Features

-   **Server:** Listens for a single client connection.
-   **Client:** Connects to the server.
-   **Tunneling:** Forwards all IP traffic from the client's virtual TUN interface to the server.
-   **Internet Access:** The server can be configured to provide internet access to the client.

## Project Structure

```
vpn-in-go/
├── .gitignore
├── README.md
├── client/
│   └── main.go
└── server/
    └── main.go
```

## Prerequisites

-   Go (version 1.18 or later)
-   `root` or `sudo` privileges to create and configure network interfaces.
-   Two machines (or VMs) for the server and client.

## Dependencies

This project uses `github.com/songgao/water` to manage TUN interfaces. The Go module system will handle the installation automatically when you build the project.

```bash
go get github.com/songgao/water
```

## How to Build

1.  **Build the server:**
    ```bash
    cd vpn-in-go/server
    go build
    ```

2.  **Build the client:**
    ```bash
    cd vpn-in-go/client
    go build
    ```

## How to Run

You will need two separate terminal windows on each machine: one to run the application and another to configure the network.

### On the Server Machine

Let's assume the server's public IP is `SERVER_PUBLIC_IP`.

1.  **Run the server binary** (requires `sudo`):
    ```bash
    sudo ./server
    ```

2.  **In a separate terminal, configure the server's TUN interface:**
    ```bash
    # Assign an IP to the new TUN interface
    sudo ip addr add 10.8.0.1/24 dev vpn-tun0

    # Bring the interface up
    sudo ip link set up dev vpn-tun0
    ```

3.  **(Optional) Enable IP forwarding to route client traffic to the internet:**
    ```bash
    # Enable IP forwarding
    sudo sysctl -w net.ipv4.ip_forward=1

    # Add a NAT rule to masquerade traffic from the VPN subnet
    # Replace 'eth0' with your server's primary network interface (e.g., ens3, enp0s3)
    sudo iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -o eth0 -j MASQUERADE
    ```

### On the Client Machine

1.  **Run the client binary** (requires `sudo`), pointing it to your server's IP and port:
    ```bash
    sudo ./client SERVER_PUBLIC_IP:8888
    ```

2.  **In a separate terminal, configure the client's TUN interface:**
    ```bash
    # Assign an IP from the VPN subnet
    sudo ip addr add 10.8.0.2/24 dev vpn-tun0

    # Bring the interface up
    sudo ip link set up dev vpn-tun0
    ```

3.  **Configure routing to send traffic through the VPN:**
    ```bash
    # Add a specific route for the VPN server to go through your original gateway
    # This prevents a routing loop.
    sudo ip route add SERVER_PUBLIC_IP via $(ip route | grep default | awk '{print $3}')

    # Change the default route to point to the VPN server's TUN interface
    sudo ip route add 0.0.0.0/1 via 10.8.0.1
    sudo ip route add 128.0.0.0/1 via 10.8.0.1
    ```

## Testing the Connection

1.  **From the client machine, ping the server's internal VPN IP:**
    ```bash
    ping 10.8.0.1
    ```

2.  **If you enabled IP forwarding on the server, check your public IP:**
    ```bash
    curl ifconfig.me
    # The output should be your server's public IP address.
    ```# vpn-in-go
