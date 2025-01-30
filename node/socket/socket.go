package socket

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"reflect"
)

type CfgTcp struct {
	Host string
	Port int16
}

func (cfg *CfgTcp) Defaults() {
	if len(cfg.Host) == 0 {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 2480
	}
}

type TcpSocket struct {
	cfg    CfgTcp
	tlsCfg *tls.Config
}

func NewSocket(cfg *CfgTcp) *TcpSocket {
	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}

	sock := &TcpSocket{
		cfg:    *cfg,
		tlsCfg: &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	sock.tlsCfg.Rand = rand.Reader

	return sock
}

func (t *TcpSocket) Listen() {
	service := fmt.Sprintf("%s:%d", t.cfg.Host, t.cfg.Port)
	listener, err := tls.Listen("tcp", service, t.tlsCfg)
	if err != nil {
		log.Fatalf("[Server] Listen error: %s", err)
	}
	log.Print("[Server] listening at", service)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[Server] Connection error: %s", err)
			break
		}

		defer conn.Close()
		log.Printf("[Server] New Connection %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// @TODO: Auth or drop
	for {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			log.Printf("[Server] cmd typ error: %s", err)
			continue
		}

		newCmd, ok := cmds[b[0]]
		if !ok {
			log.Printf("[Server] cmd not found: %b", b[0])
			continue
		}
		cmd := newCmd(conn)

		err = cmd.Run()
		if err != nil {
			log.Fatalf("[%s] unhandled error: %s", getType(cmd), err)
			break
			// or continue?
		}
	}

	// buf := make([]byte, 512)
	// 	n, err = conn.Write(buf[:n])

	log.Println("server: conn: closed")
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
