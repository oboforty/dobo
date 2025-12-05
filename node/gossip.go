package node

import "net"

func (node *Node) handleGossip(conn net.Conn) {
	// Handles inter-node p2p gossip protocol for managing cluster
	defer conn.Close()

}
