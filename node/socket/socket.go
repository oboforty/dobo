package socket

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
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
type shutdownHandler func(net.Conn)

type TcpSocket struct {
	cfg            CfgTcp
	tlsCfg         *tls.Config
	handleClient   clientHandler
	handleShutdown shutdownHandler
}

func New(cfg *CfgTcp, hcli clientHandler, hshutdown shutdownHandler) (*TcpSocket, error) {
	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.PrivKey)
	if err != nil {
		slog.Error(fmt.Sprintf("[Server] loadkey error: %s", err))
		return nil, err
	}

	sock := &TcpSocket{
		cfg:            *cfg,
		tlsCfg:         &tls.Config{Certificates: []tls.Certificate{cert}},
		handleClient:   hcli,
		handleShutdown: hshutdown,
	}
	sock.tlsCfg.Rand = rand.Reader

	return sock, nil
}

func (t *TcpSocket) RunServer() {
	service := fmt.Sprintf("%s:%d", t.cfg.Host, t.cfg.Port)
	listener, err := tls.Listen("tcp", service, t.tlsCfg)
	if err != nil {
		slog.Error(fmt.Sprintf("[Server] listen error: %s", err))
	}
	slog.Info(fmt.Sprintf("[Server] listening at %s", service))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT)
	defer stop()

	// handle dispose
	go func() {
		<-ctx.Done()
		listener.Close()
		t.handleShutdown(nil)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				slog.Info(fmt.Sprintf("[Server] shutting down (%s)", err))
				return
			default:
				slog.Info(fmt.Sprintf("[Server] connection error: %s", err))
				continue
			}
		}

		defer conn.Close()
		slog.Info(fmt.Sprintf("[Server] new connection %s", conn.RemoteAddr()))
		tlscon, ok := conn.(*tls.Conn)
		if ok {
			// log.Print("ok=true")
			state := tlscon.ConnectionState()
			slog.Info(fmt.Sprintf("       %v", state))
			// for _, v := range state.PeerCertificates {
			// 	slog.Info(fmt.Sprintf("        - key: %s", x509.MarshalPKIXPublicKey(v.PublicKey))
			// }
		}

		go t.handleClient(conn)
	}
}
