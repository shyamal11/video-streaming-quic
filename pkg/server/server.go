package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
)

type ServerConfig struct {
	GenTLS   bool
	CertFile string
	KeyFile  string
	Address  string
	Port     int
}

type Server struct {
	cfg  ServerConfig
	tls  *tls.Config
	conn quic.Connection
	ctx  context.Context
}

func NewServer(cfg ServerConfig) *Server {
	server := &Server{
		cfg: cfg,
	}
	server.tls = server.getTLS()
	server.ctx = context.TODO()
	return server
}

func (s *Server) getTLS() *tls.Config {
	if s.cfg.GenTLS {
		tlsConfig, err := util.GenerateTLSConfig()
		if err != nil {
			log.Fatal(err)
		}
		return tlsConfig
	} else {
		tlsConfig, err := util.BuildTLSConfig(s.cfg.CertFile, s.cfg.KeyFile)
		if err != nil {
			log.Fatal(err)
		}
		return tlsConfig
	}
}

func (s *Server) Run() error {
	address := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
	listener, err := quic.ListenAddr(address, s.tls, nil)
	if err != nil {
		log.Printf("error listening: %s", err)
		return err
	}

	//SERVER LOOP
	for {
		log.Println("Accepting new session")
		sess, err := listener.Accept(s.ctx)
		if err != nil {
			log.Printf("error accepting: %s", err)
			return err
		}

		go s.streamHandler(sess)
	}
}

func (s *Server) streamHandler(sess quic.Connection) {
	stream, err := sess.OpenStream()
	if err != nil {
		log.Printf("[server] error opening stream: %s", err)
		return
	}
	defer stream.Close()

	file, err := os.Open("test.mp4")
	if err != nil {
		log.Printf("[server] error opening video file: %s", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, pdu.MAX_PDU_SIZE)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[server] error reading from video file: %s", err)
			return
		}

		log.Printf("[server] sending %d bytes of video data", n)

		_, err = stream.Write(buffer[:n])
		if err != nil {
			log.Printf("[server]error writing to stream: %s", err)
			return
		}
	}
	log.Printf("[server] video sent successfully")
}
