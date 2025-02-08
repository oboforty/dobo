package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"log"
)

const KEYPATH = "/home/rajmund_csombordi/obodb/"

func main() {
	cert, err := tls.LoadX509KeyPair(
		KEYPATH+"nodekey.crt",
		KEYPATH+"nodekey.key",
	)

	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", "127.0.0.1:2480", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("client: handshake: ", state.HandshakeComplete)
	log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	var cmd uint8 = 14
	err = binary.Write(conn, binary.BigEndian, cmd)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}

	// reply := make([]byte, 256)
	// n, err = conn.Read(reply)
	// log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	log.Print("client: exiting")
}
