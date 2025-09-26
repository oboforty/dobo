package socket

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
)

type CfgTcp struct {
	Host    string `toml:"host"`
	Port    int16  `toml:"port"`
	PubKey  string `toml:"public_key"`
	PrivKey string `toml:"private_key"`
	Cert    string `toml:"cert"`
}

func (cfg *CfgTcp) Defaults() {
	if len(cfg.Host) == 0 {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 2480
	}
}

type clientHandler func(net.Conn)

type TcpSocket struct {
	cfg          CfgTcp
	tlsCfg       *tls.Config
	handleClient clientHandler
}

func New(cfg *CfgTcp, hc clientHandler) (*TcpSocket, error) {
	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.PrivKey)
	if err != nil {
		log.Fatalf("[Server] loadkey error: %s", err)
		return nil, err
	}

	sock := &TcpSocket{
		cfg:          *cfg,
		tlsCfg:       &tls.Config{Certificates: []tls.Certificate{cert}},
		handleClient: hc,
	}
	sock.tlsCfg.Rand = rand.Reader

	return sock, nil
}

func (t *TcpSocket) Listen() {
	service := fmt.Sprintf("%s:%d", t.cfg.Host, t.cfg.Port)
	listener, err := tls.Listen("tcp", service, t.tlsCfg)
	if err != nil {
		log.Fatalf("[Server] listen error: %s", err)
	}
	log.Print("[Server] listening at ", service)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[Server] connection error: %s", err)
			break
		}

		defer conn.Close()
		log.Printf("[Server] new connection %s", conn.RemoteAddr())
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			// log.Print("ok=true")
			state := tlscon.ConnectionState()
			for _, v := range state.PeerCertificates {
				log.Print(x509.MarshalPKIXPublicKey(v.PublicKey))
			}
		}

		go t.handleClient(conn)
	}
}
