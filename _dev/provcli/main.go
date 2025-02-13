package main

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
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

const (
	DO_PUT = true
	DO_GET = true
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
	// for _, v := range state.PeerCertificates {
	// 	fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
	// 	fmt.Println(v.Subject)
	// }
	log.Println("client handshake: ", state.HandshakeComplete)
	// log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	var dataLength uint32
	var cmdResponse byte
	var data []byte
	var tableName string = "table1"
	b := make([]byte, 1)

	err = binary.Write(conn, binary.BigEndian, CMD_LIST_TABLES)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	// read cmd
	_, err = conn.Read(b)
	if err != nil {
		log.Fatalf("cmd response read error: %s", err)
	}
	cmdResponse = b[0]
	if cmdResponse != CMD_LIST_TABLES {
		log.Fatalf("wrong response code: %d", cmdResponse)
	}

	// read tables
	err = binary.Read(conn, binary.BigEndian, &dataLength)
	if err != nil {
		log.Fatalf("read len error: %s", err)
	}
	data = make([]byte, dataLength)
	_, err = conn.Read(data)
	if err != nil {
		log.Fatalf("read error: %s", err)
	}

	var tables []string
	err = json.Unmarshal(data, &tables)
	if err != nil {
		log.Fatalf("json error: %s", err)
	}

	println("Tables:")
	for _, table := range tables {
		println(table)
	}
	println("---------------")

	// Create table if doesn't exist yet
	if !slices.Contains(tables, tableName) {
		log.Printf("Creating table %s", tableName)

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
		// read cmd
		_, err = conn.Read(b)
		if err != nil {
			log.Fatalf("cmd response read error: %s", err)
		}
		cmdResponse = b[0]
		if cmdResponse != CMD_CREATE_TABLE {
			log.Fatalf("wrong response code: %d", cmdResponse)
		}

		// read final result config
		err = binary.Read(conn, binary.BigEndian, &dataLength)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		data = make([]byte, dataLength)
		_, err = conn.Read(data)
		if err != nil {
			log.Fatalf("read error: %s", err)
		}
		println("New table created! Config:")
		println(string(data))
		println("")
	}

	// ----------------------------------------------
	// 			Put Item
	// ----------------------------------------------
	if DO_PUT {
		err = binary.Write(conn, binary.BigEndian, CMD_PUT_ITEM)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// Send Table Name
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

		// @TODO: read cmd
		// read cmd
		_, err = conn.Read(b)
		if err != nil {
			log.Fatalf("cmd response read error: %s", err)
		}
		cmdResponse = b[0]
		if cmdResponse != CMD_PUT_ITEM {
			log.Fatalf("wrong response code: %d", cmdResponse)
		}

		// Read put response value -- disabled
		// err = binary.Read(conn, binary.BigEndian, &dataLength)
		// if err != nil {
		// 	log.Fatalf("error: %s", err)
		// }
		// data = make([]byte, dataLength)
		// _, err = conn.Read(data)
		// if err != nil {
		// 	log.Fatalf("read GET data error: %s", err)
		// }
		// log.Printf("Item Value: %s", string(data))
	}

	// ----------------------------------------------
	// 			Get Item
	// ----------------------------------------------
	if DO_GET {
		err = binary.Write(conn, binary.BigEndian, CMD_GET_ITEM)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// Send Table Name
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

		// @TODO: $ITT Get Item CMD is not being sent!!

		// Read resp cmd
		_, err = conn.Read(b)
		if err != nil {
			log.Fatalf("cmd response read error: %s", err)
		}
		cmdResponse = b[0]
		if cmdResponse != CMD_GET_ITEM {
			log.Fatalf("wrong response code: %d", cmdResponse)
		}

		// Read value
		err = binary.Read(conn, binary.BigEndian, &dataLength)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		data = make([]byte, dataLength)
		_, err = conn.Read(data)
		if err != nil {
			log.Fatalf("read GET data error: %s", err)
		}

		log.Printf("Item Value: %s", string(data))
	}

	// reply := make([]byte, 256)
	// n, err = conn.Read(reply)
	// log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	// log.Print("client: exiting")
}
