package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"slices"
)

const KEYPATH = "/home/rajmund_csombordi/obodb/"

const (
	CMD_OK  uint8 = 1
	CMD_ERR uint8 = 2

	CMD_LIST_TABLES  uint8 = 4
	CMD_CREATE_TABLE uint8 = 11
	CMD_DROP_TABLE   uint8 = 13
	CMD_GET_ITEM     uint8 = 20
	CMD_PUT_ITEM     uint8 = 21
	CMD_UPD_ITEM     uint8 = 22
	CMD_DEL_ITEM     uint8 = 23
)

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
	log.Println("client handshake: ", state.HandshakeComplete)
	// log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	err = binary.Write(conn, binary.BigEndian, CMD_LIST_TABLES)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	var dataLength uint32
	err = binary.Read(conn, binary.BigEndian, &dataLength)
	if err != nil {
		log.Fatalf("read len error: %s", err)
	}

	tablesBytes := make([]byte, dataLength)
	_, err = conn.Read(tablesBytes)
	if err != nil {
		log.Fatalf("read error: %s", err)
	}

	var tables []string
	err = json.Unmarshal(tablesBytes, &tables)
	if err != nil {
		log.Fatalf("json error: %s", err)
	}

	println("Tables:")
	for table := range tables {
		println(table)
	}
	println("---------------")

	// Create table if doesn't exist yet
	if !slices.Contains(tables, "table1") {
		log.Printf("Creating table table1")

		err = binary.Write(conn, binary.BigEndian, CMD_CREATE_TABLE)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		settings := `{"name":"table1"}`
		err = binary.Write(conn, binary.BigEndian, uint8(1))
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		err = binary.Write(conn, binary.BigEndian, uint32(len(settings)))
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		err = binary.Write(conn, binary.BigEndian, []byte(settings))
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// get reply
		err = binary.Read(conn, binary.BigEndian, &dataLength)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		newTableCfg := make([]byte, dataLength)
		_, err = conn.Read(newTableCfg)
		if err != nil {
			log.Fatalf("read error: %s", err)
		}
		println("New table created! Config:")
		println(string(newTableCfg))
		println("")
	} else {
		// Put Item CMD
		err = binary.Write(conn, binary.BigEndian, CMD_PUT_ITEM)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// Send Table Name
		var tableName string = "table1"
		err = binary.Write(conn, binary.BigEndian, uint8(len(tableName)))
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		err = binary.Write(conn, binary.BigEndian, []byte(tableName))
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// Send Key -- PartKey type is int64 by default
		err = binary.Write(conn, binary.BigEndian, uint32(8))
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		err = binary.Write(conn, binary.BigEndian, int64(123456))
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// Send Value
		var content string = `{"name":"Rajmund", "age": 350, "data":{"attr1": "asdasd", "adas": 123}, "teso": false}`
		err = binary.Write(conn, binary.BigEndian, uint32(len(content)))
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		err = binary.Write(conn, binary.BigEndian, []byte(content))
		if err != nil {
			log.Fatalf("error: %s", err)
		}

	}

	// reply := make([]byte, 256)
	// n, err = conn.Read(reply)
	// log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	// log.Print("client: exiting")
}
